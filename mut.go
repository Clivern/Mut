// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Package main is the entry point for the Mut application.
package main

import (
	"embed"
	"io/fs"

	"github.com/clivern/mut/cli"
	"github.com/clivern/mut/locale"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

//go:embed web/dist/*
var static embed.FS

//go:embed locale/locales
var localeFS embed.FS

// main is the application entry point.
func main() {
	cli.Version = version
	cli.Commit = commit
	cli.Date = date
	cli.BuiltBy = builtBy
	cli.Static = static

	// Load locales
	if sub, err := fs.Sub(localeFS, "locale/locales"); err == nil {
		err = locale.Load(sub)
		if err != nil {
			panic(err)
		}
	}

	cli.Execute()
}
