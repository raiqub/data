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

import (
	"strconv"
	"testing"
	"time"
)

func testExpiration(store Store, t *testing.T) {
	store.SetLifetime(time.Millisecond * 10)

	store.Add("v1", nil)
	store.Add("v2", nil)

	if _, err := store.Get("v1"); err != nil {
		t.Error("The value v1 was not stored")
	}
	if _, err := store.Get("v2"); err != nil {
		t.Error("The value v2 was not stored")
	}

	time.Sleep(time.Millisecond * 20)

	if _, err := store.Get("v1"); err == nil {
		t.Error("The value v1 was not expired")
	}
	if _, err := store.Get("v2"); err == nil {
		t.Error("The value v2 was not expired")
	}

	if err := store.Delete("v1"); err == nil {
		t.Error("The expired value v1 should not be removable")
	}
	if err := store.Set("v2", nil); err == nil {
		t.Error("The expired value v2 should not be changeable")
	}
}

func testValueHandling(store Store, t *testing.T) {
	testValues := map[string]int{
		"v1":  3,
		"v2":  6,
		"v3":  83679,
		"v4":  2748,
		"v5":  54,
		"v6":  6,
		"v7":  2,
		"v8":  8,
		"v9":  7,
		"v10": 8,
	}
	rmTestKey := "v5"
	changeValues := map[string]int{
		"v4": 5062,
		"v9": 4099,
	}

	store.SetLifetime(time.Millisecond * 10)

	for k, v := range testValues {
		err := store.Add(k, v)
		if err != nil {
			t.Errorf("The value %s could not be added", k)
		}
	}

	count, err := store.Count()
	if err == nil && count != len(testValues) {
		t.Error("The values count do not match")
	}

	for k, v := range testValues {
		v2, err := store.Get(k)
		if err != nil {
			t.Errorf("The value %s could not be read", k)
		}
		if v2 != v {
			t.Errorf("The value %s was stored incorrectly", k)
		}
	}

	if err := store.Delete(rmTestKey); err != nil {
		t.Errorf("The value %s could not be removed", rmTestKey)
	}
	if _, err := store.Get(rmTestKey); err == nil {
		t.Errorf("The removed value %s should not be retrieved", rmTestKey)
	}
	count, err = store.Count()
	if err == nil && count == len(testValues) {
		t.Error("The values count should not match")
	}

	for k, v := range changeValues {
		err := store.Set(k, v)
		if err != nil {
			t.Errorf("The value %s could not be changed", k)
		}
	}
	for k, v := range changeValues {
		v2, err := store.Get(k)
		if err != nil {
			t.Errorf("The value %s could not be read", k)
		}
		if v2 != v {
			t.Errorf("The value %s was not changed", k)
		}
	}
}

func testKeyCollision(store Store, t *testing.T) {
	store.SetLifetime(time.Millisecond)

	if err := store.Add("v1", nil); err != nil {
		t.Error("The value v1 could not be stored")
	}
	if err := store.Add("v1", nil); err == nil {
		t.Error("The duplicated v1 could be stored")
	}
}

func testSetExpiration(store Store, t *testing.T) {
	store.SetLifetime(time.Millisecond)

	store.Add("v1", nil)
	store.SetLifetime(time.Second)
	store.Set("v1", nil)

	time.Sleep(time.Millisecond * 10)

	if _, err := store.Get("v1"); err != nil {
		t.Error("The value v1 is expired before expected")
	}
}

func benchmarkValueCreation(store Store, b *testing.B) {
	store.SetLifetime(time.Millisecond)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.Add(strconv.Itoa(i), i)
	}

	for i := 0; i < b.N; i++ {
		store.Get(strconv.Itoa(i))
	}
}
