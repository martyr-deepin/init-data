all: bin/importer1
	./bin/importer -i data -server http://repository.api.deepin.test

bin/importer1:
	GOPATH=`pwd`:`pwd`/vendor go build -o bin/importer importer


