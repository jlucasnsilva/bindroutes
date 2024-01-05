# Overview

Bindroutes is an experimental package to allow you to write controllers as structs as
such:

```go
Controller struct {
    bindroutes.Group `handle:"/users"`

    NewUser http.HandlerFunc `handle:"POST /"`
    GetUser http.HandlerFunc `handle:"GET /{id}"`
}
```

This package doesn't provide the routing capabilities, for that you will need another
like 'chi' or 'http.HandleFunc'.
