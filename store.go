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
	"reflect"
	"time"
)

// A Store represents a data store whose its stored values expires after
// specific elapsed time since its creation or last access.
type Store interface {
	// Add adds a new key:value to current store.
	//
	// Errors:
	// DuplicatedKeyError when requested key already exists.
	Add(key string, value interface{}) error

	// Count gets the number of stored values by current instance.
	//
	// Errors:
	// NotSupportedError when current method cannot be implemented.
	Count() (int, error)

	// Delete deletes the specified value.
	//
	// Errors:
	// InvalidKeyError when requested key could not be found.
	Delete(key string) error

	// Flush deletes any cached value into current instance.
	//
	// Errors:
	// NotSupportedError when current method cannot be implemented.
	Flush() error

	// Get gets the value stored by specified key and stores the result in the
	// value pointed to by ref.
	//
	// Errors:
	// InvalidKeyError when requested key could not be found.
	Get(key string, ref interface{}) error

	// GC garbage collects all expired data.
	GC()

	// Set sets the value of specified key.
	//
	// Errors:
	// InvalidKeyError when requested key could not be found.
	Set(key string, value interface{}) error

	// SetLifetime modifies the lifetime for a especified scope.
	//
	// Errors:
	// NotSupportedError when current method cannot be implemented.
	SetLifetime(time.Duration, LifetimeScope) error

	// SetTransient defines whether should extends expiration of stored value
	// when it is read or written.
	SetTransient(bool)
}

// A LifetimeScope its a value which defines scope to apply new lifetime value.
type LifetimeScope int

const (
	// ScopeAll defines that the new lifetime value should be applied for new
	// and existing store items.
	ScopeAll = LifetimeScope(0)

	// ScopeNewAndUpdated defines that new lifetime value should be applied for
	// new and updated store items.
	// A store item is updated when it is read or written.
	ScopeNewAndUpdated = LifetimeScope(1)

	// ScopeNew defines that new lifetime value should be applied for new store
	// items.
	ScopeNew = LifetimeScope(2)
)

func setValue(src, dst interface{}) error {
	if src == nil {
		return nil
	}

	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
		return &IndereferenceError{reflect.TypeOf(dst)}
	}

	dstVal.Elem().Set(srcVal)
	return nil
}
