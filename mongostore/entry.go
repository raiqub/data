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

import "time"

// A entry represents a document stored on MongoDB collection.
type entry struct {
	CreatedAt time.Time `bson:"at"`
	Key       string    `bson:"_id"`
	Value     *string   `bson:"val,omitempty"`
	IntVal    *int      `bson:"ival,omitempty"`
}

// IsExpired returns whether current value is expired.
func (d *entry) IsExpired(lifetime time.Duration) bool {
	return time.Now().After(d.CreatedAt.Add(lifetime))
}
