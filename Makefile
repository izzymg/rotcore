ROTCORE = rotcore
STREAMER = streamer
KBM = kbm
export BINDIR = ../bin

.PHONY: all clean rmobj $(STREAMER) $(KBM) $(ROTCORE)

all: $(STREAMER) $(KBM) $(ROTCORE) rmobj

$(STREAMER):
	$(MAKE) -C $(STREAMER)

$(KBM):
	$(MAKE) -C $(KBM)

$(ROTCORE):
	go build -o bin/rotcore

clean:
	$(RM) -r bin/*

rmobj:
	$(RM) bin/*.o