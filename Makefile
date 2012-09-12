PROG:=activities
GO_SRC:=$(wildcard *.go)

all: $(PROG)

$(PROG): $(GO_SRC)
	go build -o $@

clean:
	$(RM) $(PROG)
