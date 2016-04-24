dev:
	mkdir -p dist
	go run main.go example.js dist
	go run main.go -css example.css dist
	
test: build
	./minish example.js dist
	./minish -css example.css dist
	./minish example.js,second.js dist/bundle.js

build: 
	go build -o minish
	cp minish ~/bin/