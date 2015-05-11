dev:
	mkdir -p dist
	go run main.go example.js dist

build: 
	go build
	cp minish ~/bin/