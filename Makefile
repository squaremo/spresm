.PHONY: all spresm

all: spresm

spresm:
	go build -mod=readonly -o $@ ./cmd/spresm
