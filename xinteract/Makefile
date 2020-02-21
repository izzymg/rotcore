EXE = $(BINDIR)/xinteract
OBJ = $(BINDIR)/xinteract.o
SRC = xinteract.c

CC = gcc
LIBS = $(shell pkg-config --libs glib-2.0 libczmq libzmq) -lxdo
CFLAGS = -O1 -Wall $(shell pkg-config --cflags glib-2.0 libczmq libzmq)

.PHONY: all

all: $(EXE)

$(EXE): $(OBJ)
	gcc $(CFLAGS) $(LIBS) -o $@ $^

$(OBJ): $(SRC)
	gcc $(CFLAGS) -c -o $@ $^

test: xinteract_test.c
	gcc $(CFLAGS) $(LIBS) -o $@ $^