EXE = $(BINDIR)/kbm

CARGO = cargo
OPTS := build --release --bins --target-dir $(EXE)

.PHONY: all

all: $(EXE)

$(EXE):
	$(CARGO) $(OPTS) 