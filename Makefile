BIN := eth-tracker
BOOTSTRAP_BIN := eth-tracker-cache-bootstrap
DB_FILE := tracker_db 
BUILD_CONF := CGO_ENABLED=0 GOOS=linux GOARCH=amd64
BUILD_COMMIT := $(shell git rev-parse --short HEAD 2> /dev/null)
DEBUG := DEV=true

.PHONY: build run run-bootstrap clean clean-debug

clean:
	rm ${BIN} ${BOOTSTRAP_BIN}

clean-db:
	rm ${DB_FILE}

build:
	${BUILD_CONF} go build -ldflags="-X main.build=${BUILD_COMMIT} -s -w" -o ${BIN} cmd/service/*.go

run-bootstrap:
	${BUILD_CONF} ${DEBUG} go run cmd/bootstrap/main.go

run:
	${BUILD_CONF} ${DEBUG} go run cmd/service/*.go