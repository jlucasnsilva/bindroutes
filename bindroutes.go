package bindroutes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"
)

type (
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

	// BasePath is a annotation type so a routing group can be provided.
	// The controller struct should embed it with a handle tag that
	// contains only the routing group path/pattern.
	BasePath struct{}

	// plug maps methods (POST, GET, DELETE, etc) to a http.HandlerFunc
	// cast as a reflect.Value.
	plug map[string]reflect.Value

	handler struct {
		method      string
		path        reflect.Value
		handlerFunc reflect.Value
	}

	handlerGroups map[string][]handler
)

const (
	basePathTypeName = "BasePath"
	groupTagName     = "group"
)

var basePathType = reflect.TypeOf(BasePath{})

// Using binds all the handler funcs of each controller to a router method,
// for instance: for the given controller function
// <pre>
//
//	Post http.HandlerFunc `handler:"POST /something"`
//
// </pre>
// the result call will be r.Post("/something", controller.Post).
func UsingRouter(r Router, controllers ...any) {
	p := routerPlug(r)
	for _, c := range controllers {
		v := reflect.ValueOf(c).Elem()
		p.register(v)
	}
}

func UsingRouters(rs map[string]Router, controllers ...any) {
	gs := groupHandlerFuncs(controllers)
	for name, r := range rs {
		p := routerPlug(r)
		g, ok := gs[name]
		if ok {
			p.registerGroup(g)
		}
	}

	bs, _ := json.MarshalIndent(rs, "", "\t")
	fmt.Printf("\n\n\n\nrs = %v\n\n\n\n", string(bs))
}

func (p plug) register(v reflect.Value) {
	fields := reflect.VisibleFields(v.Type())
	bpath := basePath(fields)

	for i, f := range fields {
		if isGroupAnnotation(f) {
			continue
		}

		tag := f.Tag.Get("handle")
		if tag == "" {
			continue
		}

		method, pattern, _ := splitTag(tag)
		for k, handle := range p {
			if method == k {
				urlPath := path.Join(bpath, pattern)
				in := []reflect.Value{
					reflect.ValueOf(urlPath),
					v.FieldByIndex([]int{i}),
				}
				handle.Call(in)
			}
		}
	}
}

func (p plug) registerGroup(hs []handler) {
	for _, h := range hs {
		method, ok := p[h.method]
		if !ok {
			continue
		}
		in := []reflect.Value{
			h.path,
			h.handlerFunc,
		}
		method.Call(in)
	}
}

func isGroupAnnotation(f reflect.StructField) bool {
	return f.Name == basePathTypeName && f.Type == basePathType
}

func basePath(fields []reflect.StructField) string {
	for _, f := range fields {
		if isGroupAnnotation(f) {
			return f.Tag.Get("handle")
		}
	}
	return ""
}

func splitTag(tag string) (method, pattern string, groups string) {
	parts := strings.Split(tag, ",")
	if len(parts) < 1 {
		panic("Invalid handle definition. Method and route should be defined.")
	}

	elems := strings.Split(parts[0], " ")
	if len(elems) < 2 {
		panic("Invalid handle definition. Method and route should be defined.")
	}

	method, pattern = elems[0], elems[1]
	if !isHTTPMethod(method) {
		panic("Invalid method '" + method + "'.")
	}
	if len(parts) < 2 {
		return method, pattern, ""
	}

	gparts := strings.Split(parts[1], "=")
	if len(gparts) < 2 || gparts[0] != groupTagName || gparts[1] == "" {
		panic(
			"Invalid group declaration. The correct shape is 'groups=group_a'",
		)
	}
	return method, pattern, gparts[1]
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

func groupHandlerFuncs(controllers []any) handlerGroups {
	g := make(handlerGroups)
	for _, c := range controllers {
		v := reflect.ValueOf(c).Elem()
		fields := reflect.VisibleFields(v.Type())
		bpath := basePath(fields)
		for i, f := range fields {
			if isGroupAnnotation(f) {
				continue
			}

			tag := f.Tag.Get("handle")
			if tag == "" {
				continue
			}

			method, pattern, groupName := splitTag(tag)
			route, err := url.JoinPath(bpath, pattern)
			if err != nil {
				panic(
					"Unable to join path '" + bpath + "' + '" + pattern + "': " + err.Error(),
				)
			}
			g.add(groupName, method, route, v.FieldByIndex([]int{i}))
		}
	}
	return g
}

func (g handlerGroups) add(key, method, path string, h reflect.Value) {
	hs := g[key]
	if hs == nil {
		hs = make([]handler, 0, 30)
	}
	hs = append(hs, handler{
		handlerFunc: h,
		method:      method,
		path:        reflect.ValueOf(path),
	})
	g[key] = hs
}

func routerPlug(r Router) plug {
	return plug{
		"DELETE":  reflect.ValueOf(r.Delete),
		"GET":     reflect.ValueOf(r.Get),
		"HEAD":    reflect.ValueOf(r.Head),
		"OPTIONS": reflect.ValueOf(r.Options),
		"PATCH":   reflect.ValueOf(r.Patch),
		"POST":    reflect.ValueOf(r.Post),
		"PUT":     reflect.ValueOf(r.Put),
	}
}
