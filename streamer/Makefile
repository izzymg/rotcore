EXE = streamer
OBJ = streamer.o
SRC = streamer.c

CC = gcc
LIBS = $(shell pkg-config --libs glib-2.0 gstreamer-1.0 gstreamer-plugins-base-1.0)
CFLAGS = -O1 -Wall $(shell pkg-config --cflags glib-2.0 gstreamer-1.0 gstreamer-plugins-base-1.0)

.PHONY: all

all: $(EXE)

$(EXE): $(OBJ)
	gcc $(CFLAGS) -o $@ $^ $(LIBS) 

$(OBJ): $(SRC)
	gcc $(CFLAGS) -c -o $@ $^
