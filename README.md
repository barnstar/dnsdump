# DNS Dump

Just a simple DNS proxy that runs on the loopback interface to
intercept and dump DNS queries and responses.

## Usage

First set up looback IP thusly:
```bash
$ sudo ifconfig lo0 alias 127.0.1.2 up
```

Then start the proxy:
```bash
$ sudo go run proxy.go -l 127.0.1.2 -f 8.8.8.8
Listening on 127.0.1.2:53 (UDP). Forwarding to 8.8.8.8:53
Listening on 127.0.1.2:53 (TCP). Forwarding to 8.8.8.8:53
...
```

You may specify the forward DNS server with -f and the loopback
address to listen on with -l

Now you can proxy your queries:
```bash
$ dig +tcp @127.0.1.2 example.com
...
```

Or set your interface resolver to 127.0.1.2 to proxy all 
requests.





