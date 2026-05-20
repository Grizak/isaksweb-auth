BUILD_DIR = build
TARGET = $(BUILD_DIR)/server
SOURCES = src/main.go

.PHONY: all clean dev prod

all:
	go build -o $(TARGET) $(SOURCES)

dev:
	go run $(SOURCES)

prod: all
	GIN_MODE=release $(TARGET)

clean:
	rm -rf $(BUILD_DIR)