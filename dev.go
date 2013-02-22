// +build !prod

package main

import "github.com/alexcrichton/go-paste"
import _ "github.com/alexcrichton/go-paste/jsmin"
import _ "github.com/alexcrichton/go-paste/sass"
import _ "github.com/alexcrichton/go-paste/image"
import _ "net/http/pprof"

var PasteServer = paste.FileServer(paste.Config{
  Root: "./assets",
  TempDir: "./tmp",
  Version: "1.4",
})
