DEST = kgbpuzzlehunt.club.cc.cmu.edu:pirates

all: puzzlehunt
.PHONY: puzzlehunt

puzzlehunt:
	go build -v

run: puzzlehunt
	./puzzlehunt

deploy: puzzlehunt
	./puzzlehunt precompile
	CGO_ENABLED=0 GOOS=linux go build -tags prod -v
	rsync -auh --delete ./precompiled ./puzzlehunt ./templates $(DEST)
	@rm -rf precompiled
