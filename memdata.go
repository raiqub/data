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
	"time"
)

// A MemData represents a in-memory value whereof has a defined lifetime.
type MemData struct {
	expireAt time.Time
	lifetime time.Duration
	value    interface{}
}

// NewMemData creates a new in-memory data.
func NewMemData(lifetime time.Duration, value interface{}) *MemData {
	return &MemData{
		expireAt: time.Now().Add(lifetime),
		lifetime: lifetime,
		value:    value,
	}
}

// Delete removes current data.
func (i *MemData) Delete() {
	i.value = nil
}

// IsExpired returns whether current value is expired.
func (i *MemData) IsExpired() bool {
	return time.Now().After(i.expireAt)
}

// Hit postpone data expiration time to current time added to its lifetime
// duration.
func (i *MemData) Hit() {
	i.expireAt = time.Now().Add(i.lifetime)
}

// Value of current instance.
func (i *MemData) Value() interface{} {
	return i.value
}

// SetLifetime sets the lifetime duration for current instance.
func (i *MemData) SetLifetime(d time.Duration) {
	i.lifetime = d
}

// SetValue sets the value of current instance.
func (i *MemData) SetValue(value interface{}) {
	i.value = value
}
