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

package mongostore

import (
	"strconv"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/raiqub/data.v0"
	"gopkg.in/raiqub/dot.v1"
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

// A Store provides a MongoDB-backed key:value cache that expires after defined
// duration of time.
//
// It is a implementation of Store interface.
type Store struct {
	col            *mgo.Collection
	lifetime       time.Duration
	isTransient    bool
	ensureAccuracy bool
}

// New creates a new instance of MongoStore and defines the lifetime whether it
// is not already defined. The stored items lifetime are renewed when it is read
// or written.
func New(db *mgo.Database, name string, d time.Duration) *Store {
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

	return &Store{
		col,
		d,
		false,
		false,
	}
}

// Add adds a new key:value to current store.
//
// Errors
//
// dot.DuplicatedKeyError when requested key already exists.
//
// mgo.LastError when a error from MongoDB is triggered.
func (s *Store) Add(key string, value interface{}) error {
	doc := entry{
		time.Now(),
		key,
		nil,
		nil,
	}

	switch t := value.(type) {
	case int:
		doc.IntVal = &t
	case *int:
		doc.IntVal = t
	case string:
		doc.Value = &t
	case *string:
		doc.Value = t
	default:
		b, err := msgpack.Marshal(value)
		if err != nil {
			return err
		}
		strValue := string(b)
		doc.Value = &strValue
	}

	if err := s.col.Insert(&doc); err != nil {
		mgoerr := err.(*mgo.LastError)
		if mgoerr.Code == MongoDupKeyErrorCode {
			return dot.DuplicatedKeyError(key)
		}

		return err
	}

	return nil
}

func (s *Store) atomicInteger(key string, inc int) (int, error) {
	query := bson.M{"$inc": bson.M{"ival": inc}}
	if s.isTransient {
		query["$setOnInsert"] = bson.M{"at": time.Now()}
	} else {
		query["$currentDate"] = bson.M{"at": true}
	}

	change := mgo.Change{
		Update:    query,
		Upsert:    true,
		ReturnNew: true,
	}

	// db.runCommand({
	// 	findAndModify: "col",
	// 	query: { _id: %key },
	// 	update: {
	// 		$inc: { val: %inc },
	// 		$setOnInsert: { at: new ISODate() }
	// 	},
	// 	new: true,
	// 	upsert: true
	// })
	doc := entry{}
	_, err := s.col.FindId(key).Apply(change, &doc)
	if err != nil {
		return 0, err
	}

	return *doc.IntVal, nil
}

// Count gets the number of stored values by current instance.
//
// Errors:
// mgo.LastError when a error from MongoDB is triggered.
func (s *Store) Count() (int, error) {
	return s.col.Count()
}

// Decrement atomically gets the value stored by specified key and
// decrements it by one. If the key does not exist, it is created.
//
// Errors:
// InvalidTypeError when the value stored at key is not integer.
func (s *Store) Decrement(key string) (int, error) {
	return s.atomicInteger(key, -1)
}

// DecrementBy atomically gets the value stored by specified key and
// decrements it by value. If the key does not exist, it is created.
//
// Errors:
// InvalidTypeError when the value stored at key is not integer.
func (s *Store) DecrementBy(key string, value int) (int, error) {
	return s.atomicInteger(key, -1*value)
}

// Delete deletes the specified value.
//
// Errors
//
// dot.InvalidKeyError when requested key already exists.
//
// mgo.LastError when a error from MongoDB is triggered.
func (s *Store) Delete(key string) error {
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
func (s *Store) EnsureAccuracy(value bool) {
	s.ensureAccuracy = value
}

// Flush deletes any cached value into current instance.
//
// Errors:
// mgo.LastError when a error from MongoDB is triggered.
func (s *Store) Flush() error {
	_, err := s.col.RemoveAll(bson.M{})
	return err
}

// GC does nothing because MongoDB automatically deletes expired data.
func (s *Store) GC() {}

// Get gets the value stored by specified key and stores the result in the
// value pointed to by ref.
//
// Errors
//
// dot.InvalidKeyError when requested key already exists.
//
// mgo.LastError when a error from MongoDB is triggered.
func (s *Store) Get(key string, ref interface{}) error {
	if s.ensureAccuracy {
		if err := s.testExpiration(key); err != nil {
			return err
		}
	}

	if !s.isTransient {
		query := bson.M{"$currentDate": bson.M{"at": true}}
		if err := s.col.UpdateId(key, query); err != nil {
			if err == mgo.ErrNotFound {
				return dot.InvalidKeyError(key)
			}
			return err
		}
	}

	doc := entry{}
	err := s.col.FindId(key).One(&doc)
	if err != nil {
		if err == mgo.ErrNotFound {
			return dot.InvalidKeyError(key)
		}
		return err
	}

	switch t := ref.(type) {
	case *int:
		if doc.IntVal == nil {
			return data.InvalidTypeError{ref}
		}
		*t = *doc.IntVal
	case *string:
		if doc.Value == nil {
			return data.InvalidTypeError{ref}
		}
		*t = *doc.Value
	default:
		if doc.Value == nil {
			return data.InvalidTypeError{ref}
		}
		err = msgpack.Unmarshal([]byte(*doc.Value), ref)
		if err != nil {
			return err
		}
	}

	return nil
}

// Increment atomically gets the value stored by specified key and
// increments it by one. If the key does not exist, it is created.
//
// Errors:
// InvalidTypeError when the value stored at key is not integer.
func (s *Store) Increment(key string) (int, error) {
	return s.atomicInteger(key, 1)
}

// IncrementBy atomically gets the value stored by specified key and
// increments it by value. If the key does not exist, it is created.
//
// Errors:
// InvalidTypeError when the value stored at key is not integer.
func (s *Store) IncrementBy(key string, value int) (int, error) {
	return s.atomicInteger(key, value)
}

// Set sets the value of specified key.
//
// Errors
//
// dot.InvalidKeyError when requested key already exists.
//
// mgo.LastError when a error from MongoDB is triggered.
func (s *Store) Set(key string, value interface{}) error {
	qSet := bson.M{}
	unset := bson.M{}
	switch t := value.(type) {
	case int:
		qSet["ival"] = t
		unset["val"] = ""
	case *int:
		qSet["ival"] = *t
		unset["val"] = ""
	case string:
		qSet["val"] = t
		unset["ival"] = ""
	case *string:
		qSet["val"] = *t
		unset["ival"] = ""
	default:
		b, err := msgpack.Marshal(value)
		if err != nil {
			return err
		}
		qSet["val"] = string(b)
		unset["ival"] = ""
	}

	query := bson.M{"$set": qSet, "$unset": unset}
	if !s.isTransient {
		query["$currentDate"] = bson.M{"at": true}
	}

	if s.ensureAccuracy {
		if err := s.testExpiration(key); err != nil {
			return err
		}
	}

	if err := s.col.UpdateId(key, query); err != nil {
		if err == mgo.ErrNotFound {
			return dot.InvalidKeyError(key)
		}
		return err
	}

	return nil
}

// SetLifetime modifies the lifetime for new and existing stored items.
//
// Errors:
// NotSupportedError when ScopeNewAndUpdate or ScopeNew is specified.
func (s *Store) SetLifetime(d time.Duration, scope data.LifetimeScope) error {
	switch scope {
	case data.ScopeAll:
		s.col.DropIndexName(indexName)

		index := mgo.Index{
			Key:         []string{timeFieldName},
			Unique:      false,
			Background:  true,
			ExpireAfter: d,
			Name:        indexName,
		}
		s.col.EnsureIndex(index)
	case data.ScopeNewAndUpdated:
		return dot.NotSupportedError("ScopeNewAndUpdated")
	case data.ScopeNew:
		return dot.NotSupportedError("ScopeNew")
	default:
		return dot.NotSupportedError(strconv.Itoa(int(scope)))
	}

	s.lifetime = d
	return nil
}

// SetTransient defines whether should extends expiration of stored value
// when it is read or written.
func (s *Store) SetTransient(value bool) {
	s.isTransient = value
}

func (s *Store) testExpiration(key string) error {
	doc := entry{}

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

var _ data.Store = (*Store)(nil)
