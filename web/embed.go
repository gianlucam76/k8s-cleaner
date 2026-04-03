// Copyright 2026 vtmocanu. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package web

import "embed"

//go:embed all:dist
var staticAssets embed.FS

// StaticFS returns the embedded filesystem with the built SPA.
func StaticFS() embed.FS {
	return staticAssets
}
