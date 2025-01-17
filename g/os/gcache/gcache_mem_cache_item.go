// Copyright 2018 gf Author(https://github.com/gf). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gf.

package gcache

import "github.com/gf/g/os/gtime"

// IsExpired checks whether <item> is expired.
func (item *memCacheItem) IsExpired() bool {
	// Note that it should use greater than or equal judgement here
	// imagining that the cache time is only 1 millisecond.
	if item.e >= gtime.Millisecond() {
		return false
	}
	return true
}
