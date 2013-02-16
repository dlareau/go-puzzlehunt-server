// +build prod

package main

import "github.com/alexcrichton/go-paste"

const AssetDigest = true
var PasteServer paste.Server

func init() {
  srv, err := paste.CompiledFileServer("./assets")
  if err != nil {
    panic(err)
  }
  PasteServer = srv
}
