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

/*
Package data defines interfaces shared by other packages that implement data
storages.

Store

Store is the inteface implemented by an object that provides a storage for
expirable values.

A Store object can manage an application context. Creating an application
context its the recommended way to avoid global variables and strict the access
to your variables to selected functions.

The lifetime for new values and/or existing values can be modified calling
'SetLifetime()'. The new expiration time will be automatically updated as
specified by the scope parameter.

The expiration behaviour can be changed calling 'SetTransient()' to define
whether the lifetime of stored value is fixed (transient) or is extended when
it is read or written (non-transient).

LifetimeScope

A LifetimeScope which stored values will be affected by lifetime change.

Use 'ScopeAll' to apply the new lifetime for existing values and the ones that
will be created on the future.

Use 'ScopeNewAndUpdated' to apply the new lifetime for existing values when they
are read or written, and the ones that will be created on the future.

Use 'ScopeNew' to apply the new lifetime only for the ones that will be created
on the future.
*/
package data
