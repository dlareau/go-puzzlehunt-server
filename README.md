# Getting the code

1. [Install git](http://git-scm.com/downloads)
2. [Install go](http://golang.org/doc/install)
3. [Install bazaar](http://wiki.bazaar.canonical.com/Download) because labix is
   retarded
4. [Install mongodb](http://www.mongodb.org/downloads) for a database
3. Set up some sort of
   [$GOPATH](http://golang.org/cmd/go/#hdr-GOPATH_environment_variable)
4. Execute `go get github.com/alexcrichton/puzzlehunt`

The code is then located at `$GOPATH/src/github.com/alexcrichton/puzzlehunt`
which you may want to symlink to a better location.

# Running the code

```
# Run a database in the background somewhere
mongod &

# Build all dependencies and run a server
make run

# Visit the website!
open http://localhost:4000
```
