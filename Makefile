XSEND = xsend
XINTERACT = xinteract
ROTCORE = rotcore
export BINDIR = ../bin

.PHONY: all clean rmobj $(XSEND) $(XINTERACT) $(ROTCORE)

all: $(XSEND) $(XINTERACT) $(ROTCORE) rmobj

$(XSEND):
	$(MAKE) -C $(XSEND)

$(XINTERACT):
	$(MAKE) -C $(XINTERACT)

$(ROTCORE):
	go build -o bin/rotcore

clean:
	$(RM) bin/*

rmobj:
	$(RM) bin/*.o