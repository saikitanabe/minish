dev:
	mkdir -p dist
	go run main.go example.js dist
	go run main.go -css example.css dist
	
test: build
	./minish example.js dist
	./minish example.js,second.js dist/bundle.min.js
	./minish -css example.css dist
	./minish -css example.css dist/example-1.0.min.css

build:
	go build -o minish
	rm ~/bin/minish
	cp minish ~/bin/minish