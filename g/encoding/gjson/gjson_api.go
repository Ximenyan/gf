// Copyright 2017 gf Author(https://github.com/gf). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gf.

package gjson

import (
	"fmt"
	"time"

	"github.com/gf/g/container/gvar"
	"github.com/gf/g/os/gtime"
	"github.com/gf/g/util/gconv"
)

// Val returns the json value.
func (j *Json) Value() interface{} {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return *(j.p)
}

// Get returns value by specified <pattern>.
// It returns all values of current Json object, if <pattern> is empty or not specified.
// It returns nil if no value found by <pattern>.
//
// We can also access slice item by its index number in <pattern>,
// eg: "items.name.first", "list.10".
//
// It returns a default value specified by <def> if value for <pattern> is not found.
func (j *Json) Get(pattern string, def ...interface{}) interface{} {
	j.mu.RLock()
	defer j.mu.RUnlock()

	var result *interface{}
	if j.vc {
		result = j.getPointerByPattern(pattern)
	} else {
		result = j.getPointerByPatternWithoutViolenceCheck(pattern)
	}
	if result != nil {
		return *result
	}
	if len(def) > 0 {
		return def[0]
	}
	return nil
}

// GetVar returns a *gvar.Var with value by given <pattern>.
func (j *Json) GetVar(pattern string, def ...interface{}) *gvar.Var {
	return gvar.New(j.Get(pattern, def...), true)
}

// GetMap gets the value by specified <pattern>,
// and converts it to map[string]interface{}.
func (j *Json) GetMap(pattern string, def ...interface{}) map[string]interface{} {
	result := j.Get(pattern, def...)
	if result != nil {
		return gconv.Map(result)
	}
	return nil
}

// GetJson gets the value by specified <pattern>,
// and converts it to a un-concurrent-safe Json object.
func (j *Json) GetJson(pattern string, def ...interface{}) *Json {
	result := j.Get(pattern, def...)
	if result != nil {
		return New(result, true)
	}
	return nil
}

// GetJsons gets the value by specified <pattern>,
// and converts it to a slice of un-concurrent-safe Json object.
func (j *Json) GetJsons(pattern string, def ...interface{}) []*Json {
	array := j.GetArray(pattern, def...)
	if len(array) > 0 {
		jsonSlice := make([]*Json, len(array))
		for i := 0; i < len(array); i++ {
			jsonSlice[i] = New(array[i], true)
		}
		return jsonSlice
	}
	return nil
}

// GetJsonMap gets the value by specified <pattern>,
// and converts it to a map of un-concurrent-safe Json object.
func (j *Json) GetJsonMap(pattern string, def ...interface{}) map[string]*Json {
	m := j.GetMap(pattern, def...)
	if len(m) > 0 {
		jsonMap := make(map[string]*Json, len(m))
		for k, v := range m {
			jsonMap[k] = New(v, true)
		}
		return jsonMap
	}
	return nil
}

// GetArray gets the value by specified <pattern>,
// and converts it to a slice of []interface{}.
func (j *Json) GetArray(pattern string, def ...interface{}) []interface{} {
	return gconv.Interfaces(j.Get(pattern, def...))
}

// GetString gets the value by specified <pattern>,
// and converts it to string.
func (j *Json) GetString(pattern string, def ...interface{}) string {
	return gconv.String(j.Get(pattern, def...))
}

// GetBool gets the value by specified <pattern>,
// and converts it to bool.
// It returns false when value is: "", 0, false, off, nil;
// or returns true instead.
func (j *Json) GetBool(pattern string, def ...interface{}) bool {
	return gconv.Bool(j.Get(pattern, def...))
}

func (j *Json) GetInt(pattern string, def ...interface{}) int {
	return gconv.Int(j.Get(pattern, def...))
}

func (j *Json) GetInt8(pattern string, def ...interface{}) int8 {
	return gconv.Int8(j.Get(pattern, def...))
}

func (j *Json) GetInt16(pattern string, def ...interface{}) int16 {
	return gconv.Int16(j.Get(pattern, def...))
}

func (j *Json) GetInt32(pattern string, def ...interface{}) int32 {
	return gconv.Int32(j.Get(pattern, def...))
}

func (j *Json) GetInt64(pattern string, def ...interface{}) int64 {
	return gconv.Int64(j.Get(pattern, def...))
}

func (j *Json) GetUint(pattern string, def ...interface{}) uint {
	return gconv.Uint(j.Get(pattern, def...))
}

func (j *Json) GetUint8(pattern string, def ...interface{}) uint8 {
	return gconv.Uint8(j.Get(pattern, def...))
}

func (j *Json) GetUint16(pattern string, def ...interface{}) uint16 {
	return gconv.Uint16(j.Get(pattern, def...))
}

func (j *Json) GetUint32(pattern string, def ...interface{}) uint32 {
	return gconv.Uint32(j.Get(pattern, def...))
}

func (j *Json) GetUint64(pattern string, def ...interface{}) uint64 {
	return gconv.Uint64(j.Get(pattern, def...))
}

func (j *Json) GetFloat32(pattern string, def ...interface{}) float32 {
	return gconv.Float32(j.Get(pattern, def...))
}

func (j *Json) GetFloat64(pattern string, def ...interface{}) float64 {
	return gconv.Float64(j.Get(pattern, def...))
}

func (j *Json) GetFloats(pattern string, def ...interface{}) []float64 {
	return gconv.Floats(j.Get(pattern, def...))
}

func (j *Json) GetInts(pattern string, def ...interface{}) []int {
	return gconv.Ints(j.Get(pattern, def...))
}

// GetStrings gets the value by specified <pattern>,
// and converts it to a slice of []string.
func (j *Json) GetStrings(pattern string, def ...interface{}) []string {
	return gconv.Strings(j.Get(pattern, def...))
}

// See GetArray.
func (j *Json) GetInterfaces(pattern string, def ...interface{}) []interface{} {
	return gconv.Interfaces(j.Get(pattern, def...))
}

func (j *Json) GetTime(pattern string, format ...string) time.Time {
	return gconv.Time(j.Get(pattern), format...)
}

func (j *Json) GetDuration(pattern string, def ...interface{}) time.Duration {
	return gconv.Duration(j.Get(pattern, def...))
}

func (j *Json) GetGTime(pattern string, format ...string) *gtime.Time {
	return gconv.GTime(j.Get(pattern), format...)
}

// Set sets value with specified <pattern>.
// It supports hierarchical data access by char separator, which is '.' in default.
func (j *Json) Set(pattern string, value interface{}) error {
	return j.setValue(pattern, value, false)
}

// Remove deletes value with specified <pattern>.
// It supports hierarchical data access by char separator, which is '.' in default.
func (j *Json) Remove(pattern string) error {
	return j.setValue(pattern, nil, true)
}

// Contains checks whether the value by specified <pattern> exist.
func (j *Json) Contains(pattern string) bool {
	return j.Get(pattern) != nil
}

// Len returns the length/size of the value by specified <pattern>.
// The target value by <pattern> should be type of slice or map.
// It returns -1 if the target value is not found, or its type is invalid.
func (j *Json) Len(pattern string) int {
	p := j.getPointerByPattern(pattern)
	if p != nil {
		switch (*p).(type) {
		case map[string]interface{}:
			return len((*p).(map[string]interface{}))
		case []interface{}:
			return len((*p).([]interface{}))
		default:
			return -1
		}
	}
	return -1
}

// Append appends value to the value by specified <pattern>.
// The target value by <pattern> should be type of slice.
func (j *Json) Append(pattern string, value interface{}) error {
	p := j.getPointerByPattern(pattern)
	if p == nil {
		return j.Set(fmt.Sprintf("%s.0", pattern), value)
	}
	switch (*p).(type) {
	case []interface{}:
		return j.Set(fmt.Sprintf("%s.%d", pattern, len((*p).([]interface{}))), value)
	}
	return fmt.Errorf("invalid variable type of %s", pattern)
}

// GetToVar gets the value by specified <pattern>,
// and converts it to specified golang variable <v>.
// The <pointer> should be a pointer type.
func (j *Json) GetToVar(pattern string, pointer interface{}) error {
	r := j.Get(pattern)
	if r != nil {
		if t, err := Encode(r); err == nil {
			return DecodeTo(t, pointer)
		} else {
			return err
		}
	} else {
		pointer = nil
	}
	return nil
}

// GetToStruct gets the value by specified <pattern>,
// and converts it to specified object <objPointer>.
// The <objPointer> should be the pointer to an object.
func (j *Json) GetToStruct(pattern string, pointer interface{}) error {
	return gconv.Struct(j.Get(pattern), pointer)
}

// ToMap converts current Json object to map[string]interface{}.
// It returns nil if fails.
func (j *Json) ToMap() map[string]interface{} {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return gconv.Map(*(j.p))
}

// ToArray converts current Json object to []interface{}.
// It returns nil if fails.
func (j *Json) ToArray() []interface{} {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return gconv.Interfaces(*(j.p))
}

// ToStruct converts current Json object to specified object.
// The <objPointer> should be a pointer type.
func (j *Json) ToStruct(pointer interface{}) error {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return gconv.Struct(*(j.p), pointer)
}

// Dump prints current Json object with more manually readable.
func (j *Json) Dump() error {
	j.mu.RLock()
	defer j.mu.RUnlock()
	if b, err := j.ToJsonIndent(); err != nil {
		return err
	} else {
		fmt.Println(string(b))
	}
	return nil
}
