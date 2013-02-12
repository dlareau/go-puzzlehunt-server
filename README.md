# Running the code

1. [Install git](http://git-scm.com/downloads)
2. [Install go](http://golang.org/doc/install)
3. Set up some sort of
   [$GOPATH](http://golang.org/cmd/go/#hdr-GOPATH_environment_variable)
4. Execute `go get github.com/alexcrichton/puzzlehunt`

The code is then located at `$GOPATH/src/github.com/alexcrichton/puzzlehunt`
which you may want to symlink to a better location. Once in that directory, just
use `go build` to build and then `./puzzlehunt` to run the project
