TARGET=simple
SERVER=simple_server
CC=gcc
CFLAGS  += `pkg-config --cflags glib-2.0 gstreamer-1.0`
LDFLAGS += `pkg-config --libs glib-2.0 gstreamer-1.0`

all:$(SERVER) $(TARGET)

$(TARGET):$(TARGET).c
	$(CC) $^ -o $@ $(CFLAGS) $(LDFLAGS)

$(SERVER):$(SERVER).c
	$(CC) $^ -o $@ $(CFLAGS) $(LDFLAGS)