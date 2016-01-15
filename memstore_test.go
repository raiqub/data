/*
 * Copyright (C) 2015 Fabrício Godoy <skarllot@gmail.com>
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License
 * as published by the Free Software Foundation; either version 2
 * of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
 */

package data

import "testing"

func TestMemStore(t *testing.T) {
	store := NewCacheStore(0)
	testExpiration(store, t)

	store.Flush()
	testValueHandling(store, t)

	store.Flush()
	testKeyCollision(store, t)

	store.Flush()
	testSetExpiration(store, t)

	store.Flush()
	testPostpone(store, t)

	store.Flush()
	testTransient(store, t)
}

func BenchmarkMemStoreAddGet(b *testing.B) {
	store := NewCacheStore(0)
	benchmarkAddGet(store, b)
}

func BenchmarkMemStoreAddGetTransient(b *testing.B) {
	store := NewCacheStore(0)
	store.SetTransient(true)
	benchmarkAddGet(store, b)
}
