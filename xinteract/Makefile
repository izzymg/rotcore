EXE = $(BINDIR)/xinteract
OBJ = $(BINDIR)/xinteract.o
SRC = xinteract.c

CC = gcc
LIBS = $(shell pkg-config --libs libczmq libzmq) -lxdo
CFLAGS = -O1 -Wall $(shell pkg-config --cflags libczmq libzmq)

.PHONY: all

all: $(EXE)

$(EXE): $(OBJ)
	gcc $(CFLAGS) $(LIBS) -o $@ $^

$(OBJ): $(SRC)
	gcc $(CFLAGS) -c -o $@ $^

