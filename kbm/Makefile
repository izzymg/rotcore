EXE = $(BINDIR)/kbm

CARGO = cargo
OPTS := build --release --bins --target-dir $(EXE)

.PHONY: all $(EXE)

all: $(EXE)

$(EXE):
	$(CARGO) $(OPTS) 