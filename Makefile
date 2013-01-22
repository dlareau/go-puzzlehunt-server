DEST = kgbpuzzlehunt.club.cc.cmu.edu:pirates

all: build

build:
	go build -v

run: build
	./puzzlehunt

deploy:
	CGO_ENABLED=0 GOOS=linux go build -v
	rsync -vauh --delete puzzlehunt ./assets ./templates $(DEST)
