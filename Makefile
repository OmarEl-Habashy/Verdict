.PHONY: build run test lint clean

build:
	go build -o bin/qagent.exe .

run:
	./bin/qagent.exe --file $(FILE) --model-url $(URL) --model $(MODEL)

test:
	go test -v -count=1 ./...

lint:
	go vet ./...

clean:
	if exist bin\qagent.exe del /f bin\qagent.exe
	del /s /q testdata\*_test.go 2>nul || true
