all: bin/importer
	./bin/importer -i data -server http://repository.api.deepin.test

bin/importer:
	GOPATH=`pwd`:`pwd`/vendor go build -o bin/importer importer


