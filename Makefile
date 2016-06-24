.PHONY: broker logic http-api proxy xpush clean

all: broker logic http-api proxy xpush

broker:
	godep go build -ldflags "$(ldflags)" -o xchat-$@ xim/xchat/$@

logic:
	godep go build -ldflags "$(ldflags)" -o xchat-$@ xim/xchat/$@

http-api:
	godep go build -ldflags "$(ldflags)" -o xchat-$@ xim/xchat/$@

proxy:
	godep go build -ldflags "$(ldflags)" -o xchat-$@ xim/xchat/$@

xpush:
	godep go build -ldflags "$(ldflags)" -o xchat-$@ xim/xchat/$@

clean:
	rm -f xchat-*
