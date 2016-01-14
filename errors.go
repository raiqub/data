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
	"fmt"
	"reflect"
)

// An IndereferenceError describes a Store value that was
// not appropriate for a value of a specific Go type.
type IndereferenceError struct {
	Type reflect.Type
}

func (e *IndereferenceError) Error() string {
	if e.Type == nil {
		return "cannot unwrap to nil"
	}

	if e.Type.Kind() != reflect.Ptr {
		return fmt.Sprintf("The type %s is not a pointer", e.Type.String())
	}
	return fmt.Sprintf("Invalid type %s", e.Type.String())
}
