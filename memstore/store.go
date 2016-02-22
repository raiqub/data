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
	values      map[string]*entry
	lifetime    time.Duration
	isTransient bool
	mutex       sync.RWMutex
	gcRunning   bool
}

// New creates a new instance of in-memory Store and defines the default
// lifetime for new stored items.
//
// If it is specified to not transient then the stored items lifetime are
// renewed when it is read or written; Otherwise, it is never renewed.
func New(d time.Duration, isTransient bool) *Store {
	return &Store{
		values:      make(map[string]*entry),
		lifetime:    d,
		isTransient: isTransient,
	}
}

// Add adds a new key:value to current store.
//
// Errors:
// DuplicatedKeyError when requested key already exists.
func (s *Store) Add(key string, value interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data, err := newEntry(s.lifetime, value)
	if err != nil {
		return err
	}

	if _, ok := s.values[key]; ok {
		return dot.DuplicatedKeyError(key)
	}

	if !s.gcRunning {
		go s.gc()
	}
	s.values[key] = data
	return nil
}

func (s *Store) atomicInteger(key string, inc int) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	v, err := s.unsafeGet(key)
	if err != nil {
		data, err := newEntry(s.lifetime, inc)
		if err != nil {
			return 0, err
		}

		if !s.gcRunning {
			go s.gc()
		}
		s.values[key] = data
		return inc, nil
	}

	var value int
	if err := v.Value(&value); err != nil {
		return 0, err
	}

	value += inc
	v.SetValue(value)

	if !s.isTransient {
		v.SetLifetime(s.lifetime)
		v.Hit()
	}

	return value, nil
}

// Count gets the number of stored values by current instance.
func (s *Store) Count() (int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.values), nil
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

// Delete deletes the specified key:value.
//
// Errors:
// InvalidKeyError when requested key could not be found.
func (s *Store) Delete(key string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, err := s.unsafeGet(key)
	if err != nil {
		return err
	}

	delete(s.values, key)
	return nil
}

// Flush deletes any cached value into current instance.
func (s *Store) Flush() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.values = make(map[string]*entry)
	return nil
}

// Get gets the value stored by specified key.
//
// Errors:
// InvalidKeyError when requested key could not be found.
func (s *Store) Get(key string, ref interface{}) error {
	if s.isTransient {
		s.mutex.RLock()
		defer s.mutex.RUnlock()
	} else {
		s.mutex.Lock()
		defer s.mutex.Unlock()
	}

	v, err := s.unsafeGet(key)
	if err != nil {
		return err
	}
	if !s.isTransient {
		v.SetLifetime(s.lifetime)
		v.Hit()
	}

	return v.Value(ref)
}

func (s *Store) gc() {
	s.mutex.Lock()
	if s.gcRunning {
		s.mutex.Unlock()
		return
	}

	// Schedule GC at 1/5 intervals of current lifetime.
	interval := s.lifetime / 5
	s.gcRunning = true
	s.mutex.Unlock()

	for {
		<-time.After(interval)

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

		interval = s.lifetime / 5
		isEmpty := len(s.values) == 0
		if isEmpty {
			s.gcRunning = false
		}
		if writeLocked {
			s.mutex.Unlock()
		} else {
			s.mutex.RUnlock()
		}

		if isEmpty {
			return
		}
	}
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
// Errors:
// InvalidKeyError when requested key could not be found.
func (s *Store) Set(key string, value interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

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
	s.mutex.Lock()
	defer s.mutex.Unlock()

	switch scope {
	case data.ScopeAll:
		for _, v := range s.values {
			v.SetLifetime(d)
		}
	case data.ScopeNewAndUpdated:
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

// unsafeGet gets one entry instance from its key without locking.
//
// Errors:
// InvalidKeyError when requested key could not be found.
func (s *Store) unsafeGet(key string) (*entry, error) {
	v, ok := s.values[key]
	if !ok {
		return nil, dot.InvalidKeyError(key)
	}
	return v, nil
}

var _ data.Store = (*Store)(nil)
