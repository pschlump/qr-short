
all:
	go build

start_local:
	( cd .. ; ./qr-short & )

local_test_000:
	mkdir -p out
	( cd out ; wget 'http://127.0.0.1:2004/status' )

local_test_001:
	rm -f bulk-post2
	go build
	./bulk-post2 --data test1.csv

