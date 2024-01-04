package bindroutes

import (
	"net/http"
	"reflect"
	"strings"
)

type (
	// HandleFunc is the type of http.HandleFunc function.
	HandleFunc func(pattern string, handler func(http.ResponseWriter, *http.Request))

	// Router represents a generic HTTP router. Each of its methods
	// corresponds to one of the HTTP methods.
	Router interface {
		Delete(pattern string, h http.HandlerFunc)
		Get(pattern string, h http.HandlerFunc)
		Head(pattern string, h http.HandlerFunc)
		Options(pattern string, h http.HandlerFunc)
		Patch(pattern string, h http.HandlerFunc)
		Post(pattern string, h http.HandlerFunc)
		Put(pattern string, h http.HandlerFunc)
	}

	// plug maps methods (POST, GET, DELETE, etc) to a http.HandlerFunc
	// cast as a reflect.Value.
	plug map[string]reflect.Value
)

// Using binds all the handler funcs of each controller to a HandleFunc,
// for instance: for the given controller function
// <pre>
//
//	Post http.HandlerFunc `handler:"POST /something"`
//
// </pre>
// the result call will be fn("POST /something", controller.Post).
func Using(fn HandleFunc, controllers ...any) {
	p := plug{
		"DELETE": reflect.ValueOf(func(pattern string, h http.HandlerFunc) {
			fn("Delete "+pattern, h)
		}),
		"GET": reflect.ValueOf(func(pattern string, h http.HandlerFunc) {
			fn("Get "+pattern, h)
		}),
		"HEAD": reflect.ValueOf(func(pattern string, h http.HandlerFunc) {
			fn("Head "+pattern, h)
		}),
		"OPTIONS": reflect.ValueOf(func(pattern string, h http.HandlerFunc) {
			fn("Options "+pattern, h)
		}),
		"PATCH": reflect.ValueOf(func(pattern string, h http.HandlerFunc) {
			fn("Patch "+pattern, h)
		}),
		"POST": reflect.ValueOf(func(pattern string, h http.HandlerFunc) {
			fn("Post "+pattern, h)
		}),
		"PUT": reflect.ValueOf(func(pattern string, h http.HandlerFunc) {
			fn("Put "+pattern, h)
		}),
	}

	using(p, controllers)
}

// Using binds all the handler funcs of each controller to a router method,
// for instance: for the given controller function
// <pre>
//
//	Post http.HandlerFunc `handler:"POST /something"`
//
// </pre>
// the result call will be r.Post("/something", controller.Post).
func UsingRouter(r Router, controllers ...any) {
	p := plug{
		"DELETE":  reflect.ValueOf(r.Delete),
		"GET":     reflect.ValueOf(r.Get),
		"HEAD":    reflect.ValueOf(r.Head),
		"OPTIONS": reflect.ValueOf(r.Options),
		"PATCH":   reflect.ValueOf(r.Patch),
		"POST":    reflect.ValueOf(r.Post),
		"PUT":     reflect.ValueOf(r.Put),
	}

	using(p, controllers)
}

func using(p plug, controllers []any) {
	for _, c := range controllers {
		v := reflect.ValueOf(c).Elem()
		p.register(v)
	}
}

func (p plug) register(v reflect.Value) {
	for i, f := range reflect.VisibleFields(v.Type()) {
		tag := f.Tag.Get("handle")
		if tag == "" {
			continue
		}

		method, pattern := splitTag(tag)
		for k, handle := range p {
			if method == k {
				in := []reflect.Value{
					reflect.ValueOf(pattern),
					v.FieldByIndex([]int{i}),
				}
				handle.Call(in)
			}
		}
	}
}

func splitTag(tag string) (method, pattern string) {
	parts := strings.Split(tag, ",")
	if len(parts) < 1 {
		panic("Invalid handle definition. Method and route should be defined.")
	}

	elems := strings.Split(parts[0], " ")
	if len(elems) < 1 {
		panic("Invalid handle definition. Method and route should be defined.")
	}

	method, pattern = elems[0], elems[1]
	if !isHTTPMethod(method) {
		panic("Invalid method '" + method + "'.")
	}
	return method, pattern
}

func isHTTPMethod(m string) bool {
	return strings.EqualFold(m, "delete") ||
		strings.EqualFold(m, "get") ||
		strings.EqualFold(m, "head") ||
		strings.EqualFold(m, "options") ||
		strings.EqualFold(m, "patch") ||
		strings.EqualFold(m, "post") ||
		strings.EqualFold(m, "put")
}
