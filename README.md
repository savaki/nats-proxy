[![GoReport](https://goreportcard.com/badge/github.com/savaki/nats-proxy)](https://goreportcard.com/report/github.com/savaki/nats-proxy)
[![GoDoc](https://godoc.org/github.com/savaki/nats-proxy?status.svg)](https://godoc.org/github.com/savaki/nats-proxy)

# nats-proxy

http on the front; nats on the backend

```nats-proxy``` is a library for building http micro services on NATS and consists of 

1. an HTTP to NATS gateway that proxies HTTP requests and routes them over NATS 
2. a router that wraps an existing ```http.Handler``` to service the calls.

The aim of ```nats-proxy``` is to take the pain out of service discovery while at the
same time preserving the existing tooling around ```http.Handler```

### Why NATS

In a typical micro service environment, service discovery or how one micro service finds another
micro service can be a complicated thing.  Various tools have been developed to manage the 
service discovery issue including consul, etcd, and eureka.

Leverage NATS for service discovery, we dramatically reduce the amount of effort required for
services to discover each other.

### Status

Library is very fresh, but is already used in production so some changes are likely, but probably not
big ones.

### Example Usage

#### nats-proxy gateway

```go
nc, _ := nats.Connect(nats.DefaultURL)
gw, _ := nats_proxy.Wrap(h, 
  nats_proxy.WithNats(nc),
  nats_proxy.WithSubject("api"), // gateway root e.g. /
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
  nats_proxy.WithNats(nc),
  nats_proxy.WithSubject("api.sample"), // e.g. /api/sample
)
```

#### http client

```go
resp, _ := http.Get("http://127.0.0.1/api/sample")
```

For additional examples, see the examples package.

## NATS subject

```nats-proxy``` leverages the NATS subject to route requests.  Given a gateway with the subject ```api``` and 
the following url:

    http://localhost/a/b/c

the resuling subject would be

    api.a.b.c
    
#### Example

In this example, let's suppose we want to set up an api gateway using nats-proxy backed by two NATS micro services, 
Foo and Bar with routes as follows:

    / - root path
        /foo - handled by Foo service
        /bar - handled by Bar service

We can define this topology in ```nats-proxy``` with the following subjects:

* ```api``` - gateway subject, represents the root of the tree
* ```api.foo``` - subject for Foo 
* ```api.bar``` - subject for Bar 

## Running Multiple Gateways

Let's suppose we would like to run multiple gateways in using a single NATS cluster.  We might want to do this
to handle different environments or different products.  We can do this easily with ```nats-proxy``` by specifying
a different root subject for the gateway.

For example:

```go
nc, _ := nats.Connect(nats.DefaultURL)
staging, _ := nats_proxy.Wrap(h, 
  nats_proxy.Nats(nc),
  nats_proxy.Subject("api.staging"), // root subject for our staging environment 
)
```

