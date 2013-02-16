// +build !prod

package main

import "github.com/alexcrichton/go-paste"
import _ "github.com/alexcrichton/go-paste/jsmin"
import _ "github.com/alexcrichton/go-paste/sass"

const AssetDigest = false
var PasteServer = paste.FileServer("./assets", "1.0")
