bench:
	go test -bench=. -v -benchmem ./...

test:
	go test ./... -v

