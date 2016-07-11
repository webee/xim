.PHONY: broker logic http-api proxy clean

all: broker logic http-api proxy

broker:
	godep go build -ldflags "$(ldflags)" -o xchat-$@ xim/xchat/$@

logic:
	godep go build -ldflags "$(ldflags)" -o xchat-$@ xim/xchat/$@

http-api:
	godep go build -ldflags "$(ldflags)" -o xchat-$@ xim/xchat/$@

proxy:
	godep go build -ldflags "$(ldflags)" -o xchat-$@ xim/xchat/$@

clean:
	rm -f xchat-*
