# nats-proxy

http on the front; nats on the backend

The goal of nats-proxy is to allow you to use your existing web tools with as 
little change as possible.

```

browser  ---- http ---->  *nats.Proxy  ---- nats ---->  *nats.Router  -->  http.Handler   

```

### Example Usage

#### nats-proxy gateway

```go
nc, _ := nats.Connect(nats.DefaultURL)
gw, _ := nats_proxy.Wrap(h, 
  nats_proxy.Nats(nc),
  nats_proxy.Subject("api"), // gateway root e.g. /
)

```

#### service connected to NATS via nats-proxy:

Since ```nats-proxy``` wraps ```http.Handler```, existing web frameworks are 
compatible.

Let's mount an api to the path /api/sample.  

```go
h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
  io.WriteString(w, "hello world")
})

nc, _ := nats.Connect(nats.DefaultURL)
nats_proxy.Wrap(h, 
  nats_proxy.Nats(nc),
  nats_proxy.Subject("api.sample"), // e.g. /api/sample
)
```

#### http client

```go
resp, _ := http.Get("http://127.0.0.1/api/sample")
```
