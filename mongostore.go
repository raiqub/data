/*
 * Copyright 2016 Fabrício Godoy
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package data

import (
	"time"

	"github.com/raiqub/dot"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/vmihailenco/msgpack.v2"
)

const (
	indexName     = "expire_index"
	keyFieldName  = "_id"
	timeFieldName = "at"

	// MongoDupKeyErrorCode defines MongoDB error code when trying to insert a
	// duplicated key.
	MongoDupKeyErrorCode = 11000
)

type mongoData struct {
	CreatedAt time.Time `bson:"at"`
	Key       string    `bson:"_id"`
	Value     string    `bson:"val"`
}

// IsExpired returns whether current value is expired.
func (d *mongoData) IsExpired(lifetime time.Duration) bool {
	return time.Now().After(d.CreatedAt.Add(lifetime))
}

// A MongoStore provides a MongoDB-backed key:value cache that expires after
// defined duration of time.
//
// It is a implementation of Store interface.
type MongoStore struct {
	col            *mgo.Collection
	lifetime       time.Duration
	isTransient    bool
	ensureAccuracy bool
}

// NewMongoStore creates a new instance of MongoStore and defines the lifetime
// whether it is not already defined. The cached items lifetime are renewed when
// it is read or written.
func NewMongoStore(db *mgo.Database, name string, d time.Duration) *MongoStore {
	col := db.C(name)
	index := mgo.Index{
		Key:         []string{timeFieldName},
		Unique:      false,
		Background:  true,
		ExpireAfter: d,
		Name:        indexName,
	}
	err := col.EnsureIndex(index)
	if err != nil {
		return nil
	}

	return &MongoStore{
		col,
		d,
		false,
		false,
	}
}

// Add adds a new key:value to current store.
//
// Errors:
// 	- dot.DuplicatedKeyError when requested key already exists.
// 	- mgo.LastError when a error from MongoDB is triggered.
func (s *MongoStore) Add(key string, value interface{}) error {
	b, err := msgpack.Marshal(value)
	if err != nil {
		return err
	}

	doc := mongoData{
		time.Now(),
		key,
		string(b),
	}
	err = s.col.Insert(&doc)
	if err != nil {
		mgoerr := err.(*mgo.LastError)
		if mgoerr.Code == MongoDupKeyErrorCode {
			return dot.DuplicatedKeyError(key)
		}

		return err
	}

	return nil
}

// Count gets the number of stored values by current instance.
func (s *MongoStore) Count() (int, error) {
	return s.col.Count()
}

// Delete deletes the specified value.
//
// Errors:
// 	- dot.InvalidKeyError when requested key already exists.
// 	- mgo.LastError when a error from MongoDB is triggered.
func (s *MongoStore) Delete(key string) error {
	if s.ensureAccuracy {
		if err := s.testExpiration(key); err != nil {
			return err
		}
	}

	err := s.col.RemoveId(key)
	if err == mgo.ErrNotFound {
		return dot.InvalidKeyError(key)
	}

	return err
}

// EnsureAccuracy enables a double-check for expired values (slower). Because
// MongoDB does not garantee that expired data will be deleted immediately upon
// expiration.
func (s *MongoStore) EnsureAccuracy(value bool) {
	s.ensureAccuracy = value
}

// Flush deletes any cached value into current instance.
//
// Errors:
// 	- mgo.LastError when a error from MongoDB is triggered.
func (s *MongoStore) Flush() error {
	_, err := s.col.RemoveAll(bson.M{})
	return err
}

// GC does nothing because MongoDB automatically deletes expired data.
func (s *MongoStore) GC() {}

// Get gets the value stored by specified key and stores the result in the
// value pointed to by ref.
//
// Errors:
// 	- dot.InvalidKeyError when requested key already exists.
// 	- mgo.LastError when a error from MongoDB is triggered.
func (s *MongoStore) Get(key string, ref interface{}) error {
	doc := mongoData{
		time.Time{},
		"",
		"",
	}
	err := s.col.FindId(key).One(&doc)
	if err != nil {
		if err == mgo.ErrNotFound {
			return dot.InvalidKeyError(key)
		}
		return err
	}

	if doc.IsExpired(s.lifetime) {
		return dot.InvalidKeyError(key)
	}

	err = msgpack.Unmarshal([]byte(doc.Value), &ref)
	if err != nil {
		return err
	}

	return nil
}

// Set sets the value of specified key.
//
// Errors:
// 	- dot.InvalidKeyError when requested key already exists.
// 	- mgo.LastError when a error from MongoDB is triggered.
func (s *MongoStore) Set(key string, value interface{}) error {
	b, err := msgpack.Marshal(value)
	if err != nil {
		return err
	}

	doc := mongoData{
		time.Now(),
		key,
		string(b),
	}

	if s.ensureAccuracy {
		if err := s.testExpiration(key); err != nil {
			return err
		}
	}

	err = s.col.UpdateId(key, &doc)
	if err != nil {
		if err == mgo.ErrNotFound {
			return dot.InvalidKeyError(key)
		}
		return err
	}

	return nil
}

// SetLifetime modifies the lifetime for new and existing stored items.
//
// BUG(skarllot): Should behave as defined by interface.
func (s *MongoStore) SetLifetime(d time.Duration) {
	s.col.DropIndexName(indexName)
	s.lifetime = d

	index := mgo.Index{
		Key:         []string{timeFieldName},
		Unique:      false,
		Background:  true,
		ExpireAfter: d,
		Name:        indexName,
	}
	s.col.EnsureIndex(index)
}

// SetTransient defines whether should extends expiration of stored value
// when it is read or written.
func (s *MongoStore) SetTransient(value bool) {
	s.isTransient = value
}

func (s *MongoStore) testExpiration(key string) error {
	doc := mongoData{
		time.Time{},
		key,
		"",
	}

	err := s.col.FindId(key).One(&doc)
	if err != nil {
		if err == mgo.ErrNotFound {
			return dot.InvalidKeyError(key)
		}
		return err
	}
	if doc.IsExpired(s.lifetime) {
		return dot.InvalidKeyError(key)
	}

	return nil
}

var _ Store = (*MongoStore)(nil)
