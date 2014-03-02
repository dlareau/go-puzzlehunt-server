DEST = jlareau@kgbpuzzlehunt.club.cc.cmu.edu:zombies

all: puzzlehunt
.PHONY: puzzlehunt

puzzlehunt:
	go build -v

run: puzzlehunt
	./puzzlehunt

prod: puzzlehunt
	./puzzlehunt precompile
	go build -tags prod -v
	./puzzlehunt

deploy: puzzlehunt
	./puzzlehunt precompile
	CGO_ENABLED=0 GOOS=linux go build -tags prod -v
	rsync -auh --delete ./precompiled ./puzzlehunt ./templates $(DEST)
	@rm -rf precompiled
