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

package memdata

import (
	"testing"

	"github.com/raiqub/data/testdata"
)

func TestMemStore(t *testing.T) {
	store := NewCacheStore(0)
	testdata.TestExpiration(store, t)

	store.Flush()
	testdata.TestValueHandling(store, t)

	store.Flush()
	testdata.TestKeyCollision(store, t)

	store.Flush()
	testdata.TestSetExpiration(store, t)

	store.Flush()
	testdata.TestPostpone(store, t)

	store.Flush()
	testdata.TestTransient(store, t)
}

func BenchmarkMemStoreAddGet(b *testing.B) {
	store := NewCacheStore(0)
	testdata.BenchmarkAddGet(store, b)
}

func BenchmarkMemStoreAddGetTransient(b *testing.B) {
	store := NewCacheStore(0)
	store.SetTransient(true)
	testdata.BenchmarkAddGet(store, b)
}
