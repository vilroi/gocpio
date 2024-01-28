all: check build

check: 
	go vet ./...

build:
	go build -o gocpio cmd/*.go

test: all
	./gocpio *.cpio

clean:
	rm gocpio
