// +build !prod

package main

import "github.com/alexcrichton/go-paste"
import _ "github.com/alexcrichton/go-paste/jsmin"
import _ "github.com/alexcrichton/go-paste/sass"

var PasteServer = paste.FileServer(paste.Config{
  Root: "./assets",
  TempDir: "./tmp",
  Version: "1.4",
})
