// Copyright 2019 gf Author(https://github.com/gf). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gf.

package g

import "github.com/gf/g/net/ghttp"

// SetServerGraceful enables/disables graceful reload feature of http Web Server.
// This feature is disabled in default.
func SetServerGraceful(enabled bool) {
	ghttp.SetGraceful(enabled)
}
