TARGET=simple
CC=gcc
CFLAGS  += `pkg-config --cflags glib-2.0 gstreamer-1.0`
LDFLAGS += `pkg-config --libs glib-2.0 gstreamer-1.0`

all:$(TARGET)

$(TARGET):$(TARGET).c
	$(CC) $^ -o $@ $(CFLAGS) $(LDFLAGS)
