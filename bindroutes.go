package bindroutes

import (
	"net/http"
	"path"
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

	// Group is a annotation type so a routing group can be provided.
	// The controller struct should embed it with a handle tag that
	// contains only the routing group path/pattern.
	Group struct{}

	// plug maps methods (POST, GET, DELETE, etc) to a http.HandlerFunc
	// cast as a reflect.Value.
	plug map[string]reflect.Value
)

const groupTypeName = "Group"

var groupType = reflect.TypeOf(Group{})

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
	fields := reflect.VisibleFields(v.Type())
	group := routingGroup(fields)

	for i, f := range fields {
		if isGroupAnnotation(f) {
			continue
		}

		tag := f.Tag.Get("handle")
		if tag == "" {
			continue
		}

		method, pattern := splitTag(tag)
		for k, handle := range p {
			if method == k {
				p := path.Join(group, pattern)
				in := []reflect.Value{
					reflect.ValueOf(p),
					v.FieldByIndex([]int{i}),
				}
				handle.Call(in)
			}
		}
	}
}

func isGroupAnnotation(f reflect.StructField) bool {
	return f.Name == groupTypeName && f.Type == groupType
}

func routingGroup(fields []reflect.StructField) string {
	for _, f := range fields {
		if isGroupAnnotation(f) {
			return f.Tag.Get("handle")
		}
	}
	return ""
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
