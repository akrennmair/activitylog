PROG:=activities
GO_SRC:=$(wildcard *.go)

all: $(PROG) test bench

$(PROG): $(GO_SRC)
	go build -o $@

test:
	go test

bench:
	go test -bench='.*'

clean:
	$(RM) $(PROG)
