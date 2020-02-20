EXE = $(BINDIR)/xsend
OBJ = $(BINDIR)/xsend.o
SRC = xsend.c

CC = gcc
LIBS := $(shell pkg-config --libs glib-2.0 gstreamer-1.0)
CFLAGS := -O1 -Wall $(shell pkg-config --cflags glib-2.0 gstreamer-1.0)

.PHONY: all $(EXE)

all: $(EXE)

$(EXE): $(OBJ)
	gcc $(CFLAGS) $(LIBS) -o $@ $^

$(OBJ): $(SRC)
	gcc $(CFLAGS) -c -o $@ $^
