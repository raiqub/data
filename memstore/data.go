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

package memdata

import (
	"time"
)

// A Data represents a value stored in-memory.
type Data interface {
	// Delete removes current data.
	Delete()

	// IsExpired returns whether current value is expired.
	IsExpired() bool

	// Hit postpone data expiration time to current time added to its lifetime
	// duration.
	Hit()

	// Value of current instance.
	Value() interface{}

	// SetLifetime sets the lifetime duration for current instance.
	SetLifetime(time.Duration)

	// SetValue sets the value of current instance.
	SetValue(interface{})
}
