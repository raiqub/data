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

package testdata

import (
	"strconv"
	"testing"
	"time"

	"gopkg.in/raiqub/data.v0"
	"gopkg.in/raiqub/dot.v1"
)

func TestAtomic(store data.Store, t *testing.T) {
	if err := store.SetLifetime(time.Hour*1, data.ScopeAll); err != nil {
		t.Skip("Set lifetime to all items is not supported")
	}

	key := "user001"
	values := make(chan int, 5)

	go func() {
		for i := 0; i < 102; i++ {
			value, err := store.Increment(key)
			if err != nil {
				t.Errorf("Could not increment value: %v", err)
			}
			values <- value
			time.Sleep(time.Millisecond * 200)
		}
	}()
	go func() {
		for i := 0; i < 100; i++ {
			value, err := store.Decrement(key)
			if err != nil {
				t.Errorf("Could not decrement value: %v", err)
			}
			values <- value
			time.Sleep(time.Millisecond * 100)
		}
	}()

	for i := 0; i < 202; i++ {
		<-values
	}

	var value int
	if err := store.Get(key, &value); err != nil {
		t.Errorf("The value %s was not stored: %v", key, err)
	}

	if value != 2 {
		t.Errorf("The value of %s should be 2 but got %d", key, value)
	}
}

func TestExpiration(store data.Store, t *testing.T) {
	testValues := map[string]int{
		"v1": 3,
		"v2": 6,
	}

	if err := store.SetLifetime(time.Second*1, data.ScopeAll); err != nil {
		t.Skip("Set lifetime to all items is not supported")
	}

	if err := store.Add("v1", testValues["v1"]); err != nil {
		t.Errorf("Could not add value: %v", err)
	}
	if err := store.Add("v2", testValues["v2"]); err != nil {
		t.Errorf("Could not add value: %v", err)
	}
	var result int

	if err := store.Get("v1", &result); err != nil {
		t.Errorf("The value v1 was not stored: %v", err)
	}
	if err := store.Get("v2", &result); err != nil {
		t.Errorf("The value v2 was not stored: %v", err)
	}

	time.Sleep(time.Second * 3)

	err := store.Get("v1", &result)
	if _, ok := err.(dot.InvalidKeyError); !ok {
		t.Errorf("The value v1 was not expired: %v", err)
	}
	err = store.Get("v2", &result)
	if _, ok := err.(dot.InvalidKeyError); !ok {
		t.Errorf("The value v2 was not expired: %v", err)
	}

	err = store.Delete("v1")
	if _, ok := err.(dot.InvalidKeyError); !ok {
		t.Errorf("The expired value v1 should not be removable: %v", err)
	}
	err = store.Set("v2", nil)
	if _, ok := err.(dot.InvalidKeyError); !ok {
		t.Errorf("The expired value v2 should not be settable: %v", err)
	}
}

func TestPostpone(store data.Store, t *testing.T) {
	store.SetTransient(false)
	if err := store.SetLifetime(time.Second*1, data.ScopeAll); err != nil {
		t.Skip("Set lifetime to all items is not supported")
	}

	if err := store.Add("v1", 45); err != nil {
		t.Errorf("Could not add value: %v", err)
	}
	if err := store.Add("v2", 75); err != nil {
		t.Errorf("Could not add value: %v", err)
	}
	if err := store.Add("v3", 86); err != nil {
		t.Errorf("Could not add value: %v", err)
	}

	time.Sleep(time.Millisecond * 500)

	var result int
	if err := store.Get("v1", &result); err != nil {
		t.Errorf("Could not get value: %v", err)
	}
	if err := store.Set("v2", result); err != nil {
		t.Errorf("Could not set value: %v", err)
	}
	if err := store.Set("v3", result); err != nil {
		t.Errorf("Could not set value: %v", err)
	}
	if err := store.Get("v3", &result); err != nil {
		t.Errorf("Could not get value: %v", err)
	}

	time.Sleep(time.Millisecond * 600)
	if err := store.Get("v1", &result); err != nil {
		t.Errorf("Value expiration was not postponed: %v", err)
	}
	if err := store.Get("v2", &result); err != nil {
		t.Errorf("Value expiration was not postponed: %v", err)
	}
	if err := store.Get("v3", &result); err != nil {
		t.Errorf("Value expiration was not postponed: %v", err)
	}
}

func TestTransient(store data.Store, t *testing.T) {
	store.SetTransient(true)
	if err := store.SetLifetime(time.Second*1, data.ScopeAll); err != nil {
		t.Skip("Set lifetime to all items is not supported")
	}

	if err := store.Add("v1", 45); err != nil {
		t.Errorf("Could not add value: %v", err)
	}
	if err := store.Add("v2", 75); err != nil {
		t.Errorf("Could not add value: %v", err)
	}
	if err := store.Add("v3", 86); err != nil {
		t.Errorf("Could not add value: %v", err)
	}

	time.Sleep(time.Millisecond * 500)

	var result int
	if err := store.Get("v1", &result); err != nil {
		t.Errorf("Could not get value: %v", err)
	}
	if err := store.Set("v2", result); err != nil {
		t.Errorf("Could not set value: %v", err)
	}
	if err := store.Set("v3", result); err != nil {
		t.Errorf("Could not set value: %v", err)
	}
	if err := store.Get("v3", &result); err != nil {
		t.Errorf("Could not get value: %v", err)
	}

	time.Sleep(time.Millisecond * 600)
	if err := store.Get("v1", &result); err == nil {
		t.Errorf("Value expiration should not be postponed: %s", "v1")
	}
	if err := store.Get("v2", &result); err == nil {
		t.Errorf("Value expiration should not be postponed: %s", "v2")
	}
	if err := store.Get("v3", &result); err == nil {
		t.Errorf("Value expiration should not be postponed: %s", "v3")
	}
}

func TestTypeError(store data.Store, t *testing.T) {
	if err := store.SetLifetime(time.Second*1, data.ScopeAll); err != nil {
		t.Skip("Set lifetime to all items is not supported")
	}

	if err := store.Add("v1", 15); err != nil {
		t.Errorf("The value %d could not be added", 15)
	}
	var str string
	if err := store.Get("v1", &str); err == nil {
		t.Errorf("The value %s should not be read", "v1")
	}

	if err := store.Add("v2", "15"); err != nil {
		t.Errorf("The value %s could not be added", "15")
	}
	var integer int
	if err := store.Get("v2", &integer); err == nil {
		t.Errorf("The value %s should not be read", "v2")
	}
}

func TestValueHandling(store data.Store, t *testing.T) {
	type valueType struct {
		Number int
	}
	testValues := map[string]interface{}{
		"v1":  3,
		"v2":  6,
		"v3":  valueType{83679},
		"v4":  valueType{2748},
		"v5":  "lorem ipsum",
		"v6":  6.5,
		"v7":  876.49342,
		"v8":  valueType{8},
		"v9":  valueType{7},
		"v10": "raiqub",
	}
	rmTestKey := "v5"
	changeValues := map[string]valueType{
		"v10": {5062},
		"v7":  {4099},
	}

	if err := store.SetLifetime(time.Second*1, data.ScopeAll); err != nil {
		t.Skip("Set lifetime to all items is not supported")
	}

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
		var err error
		var output interface{}
		switch k {
		case "v1", "v2":
			var ref int
			err = store.Get(k, &ref)
			output = ref
		case "v3", "v4", "v8", "v9":
			var ref valueType
			err = store.Get(k, &ref)
			output = ref
		case "v5", "v10":
			var ref string
			err = store.Get(k, &ref)
			output = ref
		case "v6", "v7":
			var ref float64
			err = store.Get(k, &ref)
			output = ref
		}
		if err != nil {
			t.Errorf("The value %s could not be read", k)
		}
		if output != v {
			t.Errorf(
				"The value %s was stored incorrectly. Expected '%v' got '%v'.",
				k, v, output)
		}
	}

	var result interface{}
	if err := store.Delete(rmTestKey); err != nil {
		t.Errorf("The value %s could not be removed", rmTestKey)
	}
	if err := store.Get(rmTestKey, &result); err == nil {
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
		var v2 valueType
		err := store.Get(k, &v2)
		if err != nil {
			t.Errorf("The value %s could not be read", k)
		}
		if v2 != v {
			t.Errorf("The value %s was not changed. Expected '%v' got '%v'.",
				k, v, v2)
		}
	}
}

func TestKeyCollision(store data.Store, t *testing.T) {
	if err := store.SetLifetime(time.Millisecond, data.ScopeAll); err != nil {
		t.Skip("Set lifetime to all items is not supported")
	}

	if err := store.Add("v1", nil); err != nil {
		t.Error("The value v1 could not be stored")
	}
	err := store.Add("v1", nil)
	if _, ok := err.(dot.DuplicatedKeyError); !ok {
		t.Error("The duplicated v1 could be stored")
	}
}

func TestSetExpiration(store data.Store, t *testing.T) {
	if err := store.SetLifetime(time.Millisecond, data.ScopeAll); err != nil {
		t.Skip("Set lifetime to all items is not supported")
	}

	store.Add("v1", nil)
	if err := store.SetLifetime(time.Second, data.ScopeNewAndUpdated); err != nil {
		t.Skip("Set lifetime to new and updated items is not supported")
	}
	store.Set("v1", nil)

	time.Sleep(time.Millisecond * 10)

	var result interface{}
	if err := store.Get("v1", &result); err != nil {
		t.Error("The value v1 is expired before expected")
	}
}

func BenchmarkAddGet(store data.Store, b *testing.B) {
	if err := store.SetLifetime(time.Second*30, data.ScopeAll); err != nil {
		b.Skip("Set lifetime to all items is not supported")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := store.Add(strconv.Itoa(i), i); err != nil {
			b.Errorf("Could not add a new value: %v", err)
		}
	}

	var result int
	for i := 0; i < b.N; i++ {
		if err := store.Get(strconv.Itoa(i), &result); err != nil {
			b.Errorf("Could not get stored value: %v", err)
		}
	}

	b.StopTimer()
}

func BenchmarkAtomicIncrement(store data.Store, b *testing.B) {
	if err := store.SetLifetime(time.Second*30, data.ScopeAll); err != nil {
		b.Skip("Set lifetime to all items is not supported")
	}

	b.ResetTimer()

	b.SetParallelism(50)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := store.Increment("key001"); err != nil {
				b.Errorf("Could not increment value: %v", err)
			}
		}
	})

	b.StopTimer()

	var result int
	if err := store.Get("key001", &result); err != nil {
		b.Errorf("Could not get stored value: %v", err)
	}
	if result != b.N {
		b.Errorf("Unexpected value: got %d instead of %d", result, b.N)
	}
}
