build:
	GOPATH=`pwd`:`pwd`/vendor go build -o bin/importer importer


all: build
	./bin/importer -i data -server http://repository.api.deepin.test
