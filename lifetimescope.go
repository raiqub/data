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

// A LifetimeScope its a value which defines scope to apply new lifetime value.
type LifetimeScope int

const (
	// ScopeAll defines that the new lifetime value should be applied for new
	// and existing store items.
	ScopeAll = LifetimeScope(0)

	// ScopeNewAndUpdated defines that new lifetime value should be applied for
	// new and future updates on stored items.
	// A stored item is updated when it is read or written.
	ScopeNewAndUpdated = LifetimeScope(1)

	// ScopeNew defines that new lifetime value should be applied for new store
	// items.
	ScopeNew = LifetimeScope(2)
)
