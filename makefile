all: bin/importer1
	./bin/importer -i data -server http://repository.api.deepin.test

bin/importer1:
	GOPATH=`pwd`:`pwd`/vendor go build -o bin/importer importer

upload:
	./bin/importer -i data -server http://repository.api.deepin.test -upload=true

fix:
	./bin/importer -i data -fix=true


check-desktop:
	curl repository.api.deepin.test/metadata |jq -r ".data | .[] |.id |@text" > /tmp/app.list
	apt-file search desktop | awk -F ':' '{print $$1}' | sort | uniq > /tmp/desktop.list
	cat /tmp/app.list /tmp/desktop.list | sort | uniq -d > /tmp/has_desktop.list
	cat /tmp/has_desktop.list /tmp/app.list | sort | uniq -u
