// +build !prod

package main

import "github.com/alexcrichton/go-paste"

const AssetDigest = false
var PasteServer = paste.FileServer("./assets", "1.0")
