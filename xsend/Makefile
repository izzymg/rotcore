PROG := xsend
BINDIR := bin
SRCS := xsend.c xsend.h

CC := gcc
LIBS := $(shell pkg-config --libs glib-2.0 gstreamer-1.0)
CFLAGS := -O2 -Wall $(shell pkg-config --cflags glib-2.0 gstreamer-1.0)
CDBFLAGS := -ggdb -Og -Wall

.PHONY: all debug clean

$(PROG): $(SRCS)
	$(CC) $(CFLAGS) $^ $(LIBS) -o $(BINDIR)/$@

debug: $(SRCS)
	$(CC) $(CDBFLAGS) $^ $(LIBS) -o $(BINDIR)/$@

clean:
	$(RM) $(BINDIR)/*