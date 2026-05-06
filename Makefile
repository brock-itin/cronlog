BINARY     := cronlog
CMD_PATH   := ./cmd/cronlog
BUILD_DIR  := bin

.PHONY: all build test lint clean install

all: build

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) $(CMD_PATH)

test:
	go test ./...

test-verbose:
	go test -v ./...

lint:
	golangci-lint run ./...

clean:
	@rm -rf $(BUILD_DIR)

install: build
	cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY)

# Run a quick smoke test using the example config.
smoke: build
	./$(BUILD_DIR)/$(BINARY) \
		--config config/cronlog.example.yaml \
		--job smoke-test \
		-- echo "cronlog smoke test OK"
