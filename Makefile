dev:
	mkdir -p dist
	go run main.go example.js dist
	go run main.go -css example.css dist

build: 
	go build
	cp minish ~/bin/