PROG:=activities
GO_SRC:=$(wildcard *.go)
REVISION:=$(shell git rev-parse --short=7 HEAD)
LDFLAGS:="-X main.gitRevision $(REVISION)"

all: $(PROG)

$(PROG): $(GO_SRC)
	go build -ldflags $(LDFLAGS) -o $@

clean:
	$(RM) $(PROG)
