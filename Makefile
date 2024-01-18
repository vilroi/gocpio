all: check build

check: 
	go vet ./

build:
	go build -o gocpio *.go

test: all
	./gocpio initrd*

clean:
	rm gocpio
