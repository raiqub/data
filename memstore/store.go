/*
 * Copyright 2015 Fabrício Godoy
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

package memstore

import (
	"reflect"
	"strconv"
	"sync"
	"time"

	"gopkg.in/raiqub/data.v0"
	"gopkg.in/raiqub/dot.v1"
)

// A Store provides in-memory key:value cache that expires after defined
// duration of time.
//
// It is a implementation of Store interface.
type Store struct {
	values      map[string]Data
	lifetime    time.Duration
	isTransient bool
	mutex       sync.RWMutex
	lastgc      time.Time
}

// New creates a new instance of in-memory Store and defines the default
// lifetime for new stored items.
//
// If it is specified to not transient then the stored items lifetime are
// renewed when it is read or written; Otherwise, it is never renewed.
func New(d time.Duration, isTransient bool) *Store {
	return &Store{
		values:      make(map[string]Data),
		lifetime:    d,
		isTransient: isTransient,
	}
}

// Add adds a new key:value to current store.
//
// Errors:
// DuplicatedKeyError when requested key already exists.
func (s *Store) Add(key string, value interface{}) error {
	switch s.gc() {
	case dot.ReadLocked:
		s.mutex.RUnlock()
		s.mutex.Lock()
	case dot.Unlocked:
		s.mutex.Lock()
	}
	defer s.mutex.Unlock()

	data := NewMemData(s.lifetime, value)

	if _, ok := s.values[key]; ok {
		return dot.DuplicatedKeyError(key)
	}

	s.values[key] = data
	return nil
}

// Count gets the number of stored values by current instance.
func (s *Store) Count() (int, error) {
	switch s.gc() {
	case dot.WriteLocked:
		s.mutex.Unlock()
	case dot.ReadLocked:
		s.mutex.RUnlock()
	}

	return len(s.values), nil
}

// Delete deletes the specified key:value.
//
// Errors:
// InvalidKeyError when requested key could not be found.
func (s *Store) Delete(key string) error {
	lckStatus := s.gc()

	_, err := s.unsafeGet(key)
	if err != nil {
		switch lckStatus {
		case dot.WriteLocked:
			s.mutex.Unlock()
		case dot.ReadLocked:
			s.mutex.RUnlock()
		}

		return err
	}

	switch lckStatus {
	case dot.ReadLocked:
		s.mutex.RUnlock()
		s.mutex.Lock()
	case dot.Unlocked:
		s.mutex.Lock()
	}
	defer s.mutex.Unlock()

	delete(s.values, key)
	return nil
}

// Flush deletes any cached value into current instance.
func (s *Store) Flush() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.values = make(map[string]Data)
	return nil
}

// Get gets the value stored by specified key.
//
// Errors:
// InvalidKeyError when requested key could not be found.
func (s *Store) Get(key string, ref interface{}) error {
	switch s.gc() {
	case dot.WriteLocked:
		s.mutex.Unlock()
		s.mutex.RLock()
	case dot.Unlocked:
		s.mutex.RLock()
	}
	defer s.mutex.RUnlock()

	v, err := s.unsafeGet(key)
	if err != nil {
		return err
	}
	if !s.isTransient {
		v.SetLifetime(s.lifetime)
		v.Hit()
	}

	return setValue(v.Value(), ref)
}

// GC garbage collects all expired data.
func (s *Store) GC() {
	switch s.gc() {
	case dot.ReadLocked:
		s.mutex.RUnlock()
	case dot.WriteLocked:
		s.mutex.Unlock()
	}
}

func (s *Store) gc() dot.LockStatus {
	// Do not GC intervals smaller than 1/5 of current lifetime
	minInterval := s.lifetime / 5
	if s.lastgc.Add(minInterval).After(time.Now()) {
		return dot.Unlocked
	}

	s.lastgc = time.Now()
	writeLocked := false
	s.mutex.RLock()
	for i := range s.values {
		if s.values[i].IsExpired() {
			if !writeLocked {
				s.mutex.RUnlock()
				s.mutex.Lock()
				writeLocked = true
			}
			// TODO: Investigate how buckets are consolidated
			delete(s.values, i)
		}
	}

	if writeLocked {
		return dot.WriteLocked
	}

	return dot.ReadLocked
}

// Set sets the value of specified key.
//
// Errors:
// InvalidKeyError when requested key could not be found.
func (s *Store) Set(key string, value interface{}) error {
	switch s.gc() {
	case dot.WriteLocked:
		s.mutex.Unlock()
		s.mutex.RLock()
	case dot.Unlocked:
		s.mutex.RLock()
	}
	defer s.mutex.RUnlock()

	v, err := s.unsafeGet(key)
	if err != nil {
		return err
	}

	v.SetValue(value)

	if !s.isTransient {
		v.SetLifetime(s.lifetime)
		v.Hit()
	}
	return nil
}

// SetLifetime modifies the lifetime for new stored items or for existing items
// when it is read or written.
//
// Errors:
// NotSupportedError when ScopeNew is specified.
func (s *Store) SetLifetime(d time.Duration, scope data.LifetimeScope) error {
	switch scope {
	case data.ScopeAll:
		switch s.gc() {
		case dot.WriteLocked:
			s.mutex.Unlock()
			s.mutex.RLock()
		case dot.Unlocked:
			s.mutex.RLock()
		}
		defer s.mutex.RUnlock()

		for _, v := range s.values {
			v.SetLifetime(d)
		}
	case data.ScopeNewAndUpdated:
		s.mutex.RLock()
		defer s.mutex.RUnlock()
	case data.ScopeNew:
		return dot.NotSupportedError("ScopeNew")
	default:
		return dot.NotSupportedError(strconv.Itoa(int(scope)))
	}

	s.lifetime = d
	return nil
}

// SetTransient defines whether should extends expiration of stored value when
// it is read or written.
func (s *Store) SetTransient(value bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	s.isTransient = value
}

// unsafeGet gets one Data instance from its key without locking.
//
// Errors:
// InvalidKeyError when requested key could not be found.
func (s *Store) unsafeGet(key string) (Data, error) {
	v, ok := s.values[key]
	if !ok {
		return nil, dot.InvalidKeyError(key)
	}
	return v, nil
}

func setValue(src, dst interface{}) error {
	if src == nil {
		return nil
	}

	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
		return &data.IndereferenceError{Type: reflect.TypeOf(dst)}
	}

	dstVal.Elem().Set(srcVal)
	return nil
}

var _ data.Store = (*Store)(nil)
