# Builds all the binaries required by the room to serve to peers.

ROTCORE = rotcore
STREAMER = streamer
KBM = kbm
export BINDIR = ../deploy/bin

.PHONY: all clean rmobj $(STREAMER) $(KBM) $(ROTCORE)

all: $(STREAMER) $(KBM) $(ROTCORE) rmobj

$(STREAMER):
	$(MAKE) -C $(STREAMER)

$(KBM):
	$(MAKE) -C $(KBM)

$(ROTCORE):
	go build -o deploy/bin/rotcore

clean:
	$(RM) -r deploy/bin/*

rmobj:
	$(RM) deploy/bin/*.o