# Overview

Bindroutes is an experimental package to allow you to write controllers as structs as
such:

```go
Controller struct {
    NewUser http.HandlerFunc `handle:"POST /users"`
    GetUser http.HandlerFunc `handle:"GET /users/{id}"`
}
```

This package doesn't provide the routing capabilities, for that you will need another
like 'chi' or 'http.HandleFunc'.
