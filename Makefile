DEST = kgbpuzzlehunt.club.cc.cmu.edu:pirates

all: puzzlehunt

puzzlehunt:
	go build -v

run: puzzlehunt
	./puzzlehunt

deploy: puzzlehunt
	go build -v
	./puzzlehunt precompile
	CGO_ENABLED=0 GOOS=linux go build -tags prod -v
	rsync -vauh --delete ./precompiled ./puzzlehunt ./templates $(DEST)
	rm -rf precompiled
