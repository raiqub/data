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
	"time"

	"gopkg.in/vmihailenco/msgpack.v2"
)

// A entry represents a in-memory value managed by Store.
type entry struct {
	expireAt time.Time
	lifetime time.Duration
	value    []byte
}

// newEntry creates a new entry for Store.
func newEntry(lifetime time.Duration, value interface{}) (*entry, error) {
	b, err := msgpack.Marshal(value)
	if err != nil {
		return nil, err
	}

	return &entry{
		expireAt: time.Now().Add(lifetime),
		lifetime: lifetime,
		value:    b,
	}, nil
}

// Delete removes current data.
func (i *entry) Delete() {
	i.value = nil
}

// IsExpired returns whether current value is expired.
func (i *entry) IsExpired() bool {
	return time.Now().After(i.expireAt)
}

// Hit postpone data expiration time to current time added to its lifetime
// duration.
func (i *entry) Hit() {
	i.expireAt = time.Now().Add(i.lifetime)
}

// Value of current instance.
func (i *entry) Value(ref interface{}) error {
	err := msgpack.Unmarshal(i.value, &ref)
	if err != nil {
		return err
	}

	return nil
}

// SetLifetime sets the lifetime duration for current instance.
func (i *entry) SetLifetime(d time.Duration) {
	i.lifetime = d
}

// SetValue sets the value of current instance.
func (i *entry) SetValue(value interface{}) error {
	b, err := msgpack.Marshal(value)
	if err != nil {
		return err
	}

	i.value = b
	return nil
}
