.PHONY: all build clean run test

all: build

build: build-probe build-crawler build-parser

build-probe:
	g++ -O3 -pthread -o cmd/probe/lux_probe cmd/probe/main.cpp

build-crawler:
	cd cmd/crawler && go build -o ../lux_crawler main.go

build-parser:
	cd cmd/parser && cargo build --release && cp target/release/lux_parser ../../cmd/parser/lux_parser

clean:
	rm -f cmd/probe/lux_probe cmd/crawler/lux_crawler cmd/parser/lux_parser
	rm -f lux_crawl.db

run:
	python3 api/main.py &
	WORKER_ID=1 python3 pkg/queue/worker.py &

test:
	./cmd/probe/lux_probe localhost 1 100
