// Copyright 2018 gf Author(https://github.com/gf). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gf.

package garray

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/gf/g/container/gtype"
	"github.com/gf/g/internal/rwmutex"
	"github.com/gf/g/util/gconv"
	"github.com/gf/g/util/grand"
)

// It's using increasing order in default.
type SortedStringArray struct {
	mu         *rwmutex.RWMutex
	array      []string
	unique     *gtype.Bool             // Whether enable unique feature(false)
	comparator func(v1, v2 string) int // Comparison function(it returns -1: v1 < v2; 0: v1 == v2; 1: v1 > v2)
}

// NewSortedStringArray creates and returns an empty sorted array.
// The parameter <unsafe> used to specify whether using array in un-concurrent-safety,
// which is false in default.
func NewSortedStringArray(unsafe ...bool) *SortedStringArray {
	return NewSortedStringArraySize(0, unsafe...)
}

// NewSortedStringArraySize create and returns an sorted array with given size and cap.
// The parameter <unsafe> used to specify whether using array in un-concurrent-safety,
// which is false in default.
func NewSortedStringArraySize(cap int, unsafe ...bool) *SortedStringArray {
	return &SortedStringArray{
		mu:     rwmutex.New(unsafe...),
		array:  make([]string, 0, cap),
		unique: gtype.NewBool(),
		comparator: func(v1, v2 string) int {
			return strings.Compare(v1, v2)
		},
	}
}

// NewSortedStringArrayFrom creates and returns an sorted array with given slice <array>.
// The parameter <unsafe> used to specify whether using array in un-concurrent-safety,
// which is false in default.
func NewSortedStringArrayFrom(array []string, unsafe ...bool) *SortedStringArray {
	a := NewSortedStringArraySize(0, unsafe...)
	a.array = array
	sort.Strings(a.array)
	return a
}

// NewSortedStringArrayFromCopy creates and returns an sorted array from a copy of given slice <array>.
// The parameter <unsafe> used to specify whether using array in un-concurrent-safety,
// which is false in default.
func NewSortedStringArrayFromCopy(array []string, unsafe ...bool) *SortedStringArray {
	newArray := make([]string, len(array))
	copy(newArray, array)
	return NewSortedStringArrayFrom(newArray, unsafe...)
}

// SetArray sets the underlying slice array with the given <array>.
func (a *SortedStringArray) SetArray(array []string) *SortedStringArray {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.array = array
	sort.Strings(a.array)
	return a
}

// Sort sorts the array in increasing order.
// The parameter <reverse> controls whether sort
// in increasing order(default) or decreasing order.
func (a *SortedStringArray) Sort() *SortedStringArray {
	a.mu.Lock()
	defer a.mu.Unlock()
	sort.Strings(a.array)
	return a
}

// Add adds one or multiple values to sorted array, the array always keeps sorted.
func (a *SortedStringArray) Add(values ...string) *SortedStringArray {
	if len(values) == 0 {
		return a
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, value := range values {
		index, cmp := a.binSearch(value, false)
		if a.unique.Val() && cmp == 0 {
			continue
		}
		if index < 0 {
			a.array = append(a.array, value)
			continue
		}
		if cmp > 0 {
			index++
		}
		rear := append([]string{}, a.array[index:]...)
		a.array = append(a.array[0:index], value)
		a.array = append(a.array, rear...)
	}
	return a
}

// Get returns the value of the specified index,
// the caller should notice the boundary of the array.
func (a *SortedStringArray) Get(index int) string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	value := a.array[index]
	return value
}

// Remove removes an item by index.
func (a *SortedStringArray) Remove(index int) string {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Determine array boundaries when deleting to improve deletion efficiency.
	if index == 0 {
		value := a.array[0]
		a.array = a.array[1:]
		return value
	} else if index == len(a.array)-1 {
		value := a.array[index]
		a.array = a.array[:index]
		return value
	}
	// If it is a non-boundary delete,
	// it will involve the creation of an array,
	// then the deletion is less efficient.
	value := a.array[index]
	a.array = append(a.array[:index], a.array[index+1:]...)
	return value
}

// PopLeft pops and returns an item from the beginning of array.
func (a *SortedStringArray) PopLeft() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	value := a.array[0]
	a.array = a.array[1:]
	return value
}

// PopRight pops and returns an item from the end of array.
func (a *SortedStringArray) PopRight() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	index := len(a.array) - 1
	value := a.array[index]
	a.array = a.array[:index]
	return value
}

// PopRand randomly pops and return an item out of array.
func (a *SortedStringArray) PopRand() string {
	return a.Remove(grand.Intn(len(a.array)))
}

// PopRands randomly pops and returns <size> items out of array.
func (a *SortedStringArray) PopRands(size int) []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if size > len(a.array) {
		size = len(a.array)
	}
	array := make([]string, size)
	for i := 0; i < size; i++ {
		index := grand.Intn(len(a.array))
		array[i] = a.array[index]
		a.array = append(a.array[:index], a.array[index+1:]...)
	}
	return array
}

// PopLefts pops and returns <size> items from the beginning of array.
func (a *SortedStringArray) PopLefts(size int) []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	length := len(a.array)
	if size > length {
		size = length
	}
	value := a.array[0:size]
	a.array = a.array[size:]
	return value
}

// PopRights pops and returns <size> items from the end of array.
func (a *SortedStringArray) PopRights(size int) []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	index := len(a.array) - size
	if index < 0 {
		index = 0
	}
	value := a.array[index:]
	a.array = a.array[:index]
	return value
}

// Range picks and returns items by range, like array[start:end].
// Notice, if in concurrent-safe usage, it returns a copy of slice;
// else a pointer to the underlying data.
//
// If <end> is negative, then the offset will start from the end of array.
// If <end> is omitted, then the sequence will have everything from start up
// until the end of the array.
func (a *SortedStringArray) Range(start int, end ...int) []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	offsetEnd := len(a.array)
	if len(end) > 0 && end[0] < offsetEnd {
		offsetEnd = end[0]
	}
	if start > offsetEnd {
		return nil
	}
	if start < 0 {
		start = 0
	}
	array := ([]string)(nil)
	if a.mu.IsSafe() {
		array = make([]string, offsetEnd-start)
		copy(array, a.array[start:offsetEnd])
	} else {
		array = a.array[start:offsetEnd]
	}
	return array
}

// SubSlice returns a slice of elements from the array as specified
// by the <offset> and <size> parameters.
// If in concurrent safe usage, it returns a copy of the slice; else a pointer.
//
// If offset is non-negative, the sequence will start at that offset in the array.
// If offset is negative, the sequence will start that far from the end of the array.
//
// If length is given and is positive, then the sequence will have up to that many elements in it.
// If the array is shorter than the length, then only the available array elements will be present.
// If length is given and is negative then the sequence will stop that many elements from the end of the array.
// If it is omitted, then the sequence will have everything from offset up until the end of the array.
//
// Any possibility crossing the left border of array, it will fail.
func (a *SortedStringArray) SubSlice(offset int, length ...int) []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	size := len(a.array)
	if len(length) > 0 {
		size = length[0]
	}
	if offset > len(a.array) {
		return nil
	}
	if offset < 0 {
		offset = len(a.array) + offset
		if offset < 0 {
			return nil
		}
	}
	if size < 0 {
		offset += size
		size = -size
		if offset < 0 {
			return nil
		}
	}
	end := offset + size
	if end > len(a.array) {
		end = len(a.array)
		size = len(a.array) - offset
	}
	if a.mu.IsSafe() {
		s := make([]string, size)
		copy(s, a.array[offset:])
		return s
	} else {
		return a.array[offset:end]
	}
}

// Sum returns the sum of values in an array.
func (a *SortedStringArray) Sum() (sum int) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, v := range a.array {
		sum += gconv.Int(v)
	}
	return
}

// Len returns the length of array.
func (a *SortedStringArray) Len() int {
	a.mu.RLock()
	length := len(a.array)
	a.mu.RUnlock()
	return length
}

// Slice returns the underlying data of array.
// Notice, if in concurrent-safe usage, it returns a copy of slice;
// else a pointer to the underlying data.
func (a *SortedStringArray) Slice() []string {
	array := ([]string)(nil)
	if a.mu.IsSafe() {
		a.mu.RLock()
		defer a.mu.RUnlock()
		array = make([]string, len(a.array))
		copy(array, a.array)
	} else {
		array = a.array
	}
	return array
}

// Contains checks whether a value exists in the array.
func (a *SortedStringArray) Contains(value string) bool {
	return a.Search(value) != -1
}

// Search searches array by <value>, returns the index of <value>,
// or returns -1 if not exists.
func (a *SortedStringArray) Search(value string) (index int) {
	if i, r := a.binSearch(value, true); r == 0 {
		return i
	}
	return -1
}

// Binary search.
// It returns the last compared index and the result.
// If <result> equals to 0, it means the value at <index> is equals to <value>.
// If <result> lesser than 0, it means the value at <index> is lesser than <value>.
// If <result> greater than 0, it means the value at <index> is greater than <value>.
func (a *SortedStringArray) binSearch(value string, lock bool) (index int, result int) {
	if len(a.array) == 0 {
		return -1, -2
	}
	if lock {
		a.mu.RLock()
		defer a.mu.RUnlock()
	}
	min := 0
	max := len(a.array) - 1
	mid := 0
	cmp := -2
	for min <= max {
		mid = int((min + max) / 2)
		cmp = a.comparator(value, a.array[mid])
		switch {
		case cmp < 0:
			max = mid - 1
		case cmp > 0:
			min = mid + 1
		default:
			return mid, cmp
		}
	}
	return mid, cmp
}

// SetUnique sets unique mark to the array,
// which means it does not contain any repeated items.
// It also do unique check, remove all repeated items.
func (a *SortedStringArray) SetUnique(unique bool) *SortedStringArray {
	oldUnique := a.unique.Val()
	a.unique.Set(unique)
	if unique && oldUnique != unique {
		a.Unique()
	}
	return a
}

// Unique uniques the array, clear repeated items.
func (a *SortedStringArray) Unique() *SortedStringArray {
	a.mu.Lock()
	i := 0
	for {
		if i == len(a.array)-1 {
			break
		}
		if a.comparator(a.array[i], a.array[i+1]) == 0 {
			a.array = append(a.array[:i+1], a.array[i+1+1:]...)
		} else {
			i++
		}
	}
	a.mu.Unlock()
	return a
}

// Clone returns a new array, which is a copy of current array.
func (a *SortedStringArray) Clone() (newArray *SortedStringArray) {
	a.mu.RLock()
	array := make([]string, len(a.array))
	copy(array, a.array)
	a.mu.RUnlock()
	return NewSortedStringArrayFrom(array, !a.mu.IsSafe())
}

// Clear deletes all items of current array.
func (a *SortedStringArray) Clear() *SortedStringArray {
	a.mu.Lock()
	if len(a.array) > 0 {
		a.array = make([]string, 0)
	}
	a.mu.Unlock()
	return a
}

// LockFunc locks writing by callback function <f>.
func (a *SortedStringArray) LockFunc(f func(array []string)) *SortedStringArray {
	a.mu.Lock()
	defer a.mu.Unlock()
	f(a.array)
	return a
}

// RLockFunc locks reading by callback function <f>.
func (a *SortedStringArray) RLockFunc(f func(array []string)) *SortedStringArray {
	a.mu.RLock()
	defer a.mu.RUnlock()
	f(a.array)
	return a
}

// Merge merges <array> into current array.
// The parameter <array> can be any garray or slice type.
// The difference between Merge and Append is Append supports only specified slice type,
// but Merge supports more parameter types.
func (a *SortedStringArray) Merge(array interface{}) *SortedStringArray {
	switch v := array.(type) {
	case *Array:
		a.Add(gconv.Strings(v.Slice())...)
	case *IntArray:
		a.Add(gconv.Strings(v.Slice())...)
	case *StringArray:
		a.Add(gconv.Strings(v.Slice())...)
	case *SortedArray:
		a.Add(gconv.Strings(v.Slice())...)
	case *SortedIntArray:
		a.Add(gconv.Strings(v.Slice())...)
	case *SortedStringArray:
		a.Add(gconv.Strings(v.Slice())...)
	default:
		a.Add(gconv.Strings(array)...)
	}
	return a
}

// Chunk splits an array into multiple arrays,
// the size of each array is determined by <size>.
// The last chunk may contain less than size elements.
func (a *SortedStringArray) Chunk(size int) [][]string {
	if size < 1 {
		return nil
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	length := len(a.array)
	chunks := int(math.Ceil(float64(length) / float64(size)))
	var n [][]string
	for i, end := 0, 0; chunks > 0; chunks-- {
		end = (i + 1) * size
		if end > length {
			end = length
		}
		n = append(n, a.array[i*size:end])
		i++
	}
	return n
}

// Rand randomly returns one item from array(no deleting).
func (a *SortedStringArray) Rand() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.array[grand.Intn(len(a.array))]
}

// Rands randomly returns <size> items from array(no deleting).
func (a *SortedStringArray) Rands(size int) []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if size > len(a.array) {
		size = len(a.array)
	}
	n := make([]string, size)
	for i, v := range grand.Perm(len(a.array)) {
		n[i] = a.array[v]
		if i == size-1 {
			break
		}
	}
	return n
}

// Join joins array elements with a string <glue>.
func (a *SortedStringArray) Join(glue string) string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	buffer := bytes.NewBuffer(nil)
	for k, v := range a.array {
		buffer.WriteString(gconv.String(v))
		if k != len(a.array)-1 {
			buffer.WriteString(glue)
		}
	}
	return buffer.String()
}

// CountValues counts the number of occurrences of all values in the array.
func (a *SortedStringArray) CountValues() map[string]int {
	m := make(map[string]int)
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, v := range a.array {
		m[v]++
	}
	return m
}

// String returns current array as a string.
func (a *SortedStringArray) String() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return fmt.Sprint(a.array)
}
