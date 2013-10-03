// +build prod

package main

import "github.com/alexcrichton/go-paste"

var PasteServer paste.Server

func init() {
	srv, err := paste.CompiledFileServer("./precompiled")
	if err != nil {
		panic(err)
	}
	PasteServer = srv
}
