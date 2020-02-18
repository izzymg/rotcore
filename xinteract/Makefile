BINDIR = bin
EXE = bin/xinteract

CC = gcc
CFLAGS = -Wall -ggdb $(shell pkg-config --cflags libczmq libzmq)
LIBS = $(shell pkg-config --libs libczmq libzmq) -lxdo

.PHONY: all clean

all: $(EXE) $(PROG_SERVER)

clean:
	$(RM) $(BINDIR)/*

$(EXE): $(BINDIR)/xinteract.o
	gcc $(CFLAGS) $(LIBS) -o $@ $^

$(BINDIR)/xinteract.o: xinteract.c
	gcc $(CFLAGS) -c -o $@ $^

