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

package data

import (
	"sync"
	"time"

	"github.com/raiqub/dot"
)

// A MemStore provides in-memory key:value cache that expires after defined
// duration of time.
//
// It is a implementation of Store interface.
type MemStore struct {
	values      map[string]Data
	lifetime    time.Duration
	isTransient bool
	mutex       sync.RWMutex
}

// NewCacheStore creates a new instance of MemStore and defines the default
// lifetime for new cached items. The cached items lifetime are renewed when it
// is read or written.
func NewCacheStore(d time.Duration) *MemStore {
	return &MemStore{
		values:      make(map[string]Data),
		lifetime:    d,
		isTransient: false,
	}
}

// NewTransientStore creates a new instance of MemStore and defines the default
// lifetime for new stored items. The stored items lifetime are not renewed when
// it is read or written.
func NewTransientStore(d time.Duration) *MemStore {
	return &MemStore{
		values:      make(map[string]Data),
		lifetime:    d,
		isTransient: true,
	}
}

// Add adds a new key:value to current store.
//
// Errors:
// DuplicatedKeyError when requested key already exists.
func (s *MemStore) Add(key string, value interface{}) error {
	lckStatus := s.gc()

	data := NewMemData(s.lifetime, value)

	if lckStatus == dot.ReadLocked {
		s.mutex.RUnlock()
		s.mutex.Lock()
	}
	defer s.mutex.Unlock()

	if _, ok := s.values[key]; ok {
		return dot.DuplicatedKeyError(key)
	}

	s.values[key] = data
	return nil
}

// Count gets the number of stored values by current instance.
func (s *MemStore) Count() (int, error) {
	if s.gc() == dot.WriteLocked {
		defer s.mutex.Unlock()
	} else {
		defer s.mutex.RUnlock()
	}

	return len(s.values), nil
}

// Delete deletes the specified key:value.
//
// Errors:
// InvalidKeyError when requested key could not be found.
func (s *MemStore) Delete(key string) error {
	lckStatus := s.gc()

	_, err := s.unsafeGet(key)
	if err != nil {
		if lckStatus == dot.WriteLocked {
			s.mutex.Unlock()
		} else {
			s.mutex.RUnlock()
		}

		return err
	}

	if lckStatus == dot.ReadLocked {
		s.mutex.RUnlock()
		s.mutex.Lock()
	}
	defer s.mutex.Unlock()

	delete(s.values, key)
	return nil
}

// Flush deletes any cached value into current instance.
func (s *MemStore) Flush() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.values = make(map[string]Data)
	return nil
}

// Get gets the value stored by specified key.
//
// Errors:
// InvalidKeyError when requested key could not be found.
func (s *MemStore) Get(key string, ref interface{}) error {
	if s.gc() == dot.WriteLocked {
		s.mutex.Unlock()
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
func (s *MemStore) GC() {
	lckStatus := s.gc()

	if lckStatus == dot.ReadLocked {
		s.mutex.RUnlock()
	} else {
		s.mutex.Unlock()
	}
}

func (s *MemStore) gc() dot.LockStatus {
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
func (s *MemStore) Set(key string, value interface{}) error {
	if s.gc() == dot.WriteLocked {
		s.mutex.Unlock()
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
func (s *MemStore) SetLifetime(d time.Duration) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	s.lifetime = d
}

// SetTransient defines whether should extends expiration of stored value when
// it is read or written.
func (s *MemStore) SetTransient(value bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	s.isTransient = value
}

// unsafeGet gets one Data instance from its key without locking.
//
// Errors:
// InvalidKeyError when requested key could not be found.
func (s *MemStore) unsafeGet(key string) (Data, error) {
	v, ok := s.values[key]
	if !ok {
		return nil, dot.InvalidKeyError(key)
	}
	return v, nil
}

var _ Store = (*MemStore)(nil)
