// +build !prod

package main

import "github.com/alexcrichton/go-paste"

var PasteServer = paste.FileServer("./assets", "1.0")
