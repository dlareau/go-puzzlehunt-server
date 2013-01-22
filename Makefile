DEST = kgbpuzzlehunt.club.cc.cmu.edu:weathermen

all: build

build:
	go build -v

run: build
	./weathermen

deploy:
	CGO_ENABLED=0 GOOS=linux go build -v
	rsync -vauh --delete weathermen ./assets ./templates $(DEST)
