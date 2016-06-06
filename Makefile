.PHONY: broker logic proxy clean

all: broker logic proxy

broker:
	godep go build -ldflags "$(ldflags)" -o xchat-$@ xim/xchat/$@

logic:
	godep go build -ldflags "$(ldflags)" -o xchat-$@ xim/xchat/$@

proxy:
	godep go build -ldflags "$(ldflags)" -o xchat-$@ xim/xchat/$@

clean:
	rm -f xchat-*
