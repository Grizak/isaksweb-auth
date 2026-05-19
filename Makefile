BUILD_DIR = build
TARGET = $(BUILD_DIR)/server
SOURCES = src/main.go

.PHONY: all clean

all:
	go build -o $(TARGET) $(SOURCES)

dev: all
	$(TARGET)

clean:
	rm -rf $(BUILD_DIR)