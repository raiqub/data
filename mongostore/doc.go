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

/*
Package mongostore provides MongoDB-backed data store implementation.

Store

A Store provides MongoDB-backed key:value cache that expires after defined
duration of time. That duration is defined when a new instance is initialized
calling 'mongostore.New()' function and it is used to all new stored values.

The Store can manage an application context. Creating an application context
its the recommended way to avoid global variables and strict the access to your
variables to selected functions.

The lifetime for new values and existing values can be modified calling
'SetLifetime()'.

The expiration behaviour can be changed calling 'SetTransient()' to define
whether the lifetime of stored value is fixed (transient) or is extended when
it is read or written (non-transient).
*/
package mongostore
