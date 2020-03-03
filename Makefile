XSEND = xsend
KBM = kbm
ROTCORE = rotcore
export BINDIR = ../bin

.PHONY: all clean rmobj $(XSEND) $(KBM) $(ROTCORE)

all: $(XSEND) $(KBM) $(ROTCORE) rmobj

$(XSEND):
	$(MAKE) -C $(XSEND)

$(KBM):
	$(MAKE) -C $(KBM)

$(ROTCORE):
	go build -o bin/rotcore

clean:
	$(RM) -r bin/*

rmobj:
	$(RM) bin/*.o