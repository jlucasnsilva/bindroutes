package bindroutes

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type (
	router map[string]bool

	handler struct {
		Group `handle:"/users"`

		Post    http.HandlerFunc `handle:"POST /"`
		Get     http.HandlerFunc `handle:"GET /{id}"`
		Put     http.HandlerFunc `handle:"PUT /{id}"`
		Delete  http.HandlerFunc `handle:"DELETE /{id}"`
		Ignored string
	}

	failingHandler struct {
		Post http.HandlerFunc `handle:"Get,/users"`
	}
)

func TestRegister(t *testing.T) {
	dummy := func(w http.ResponseWriter, r *http.Request) {}
	h := handler{Post: dummy, Get: dummy, Put: dummy, Delete: dummy}
	r := make(router)

	UsingRouter(r, &h)

	assert.True(t, r["POST:/users"])
	assert.True(t, r["GET:/users/{id}"])
	assert.True(t, r["PUT:/users/{id}"])
	assert.True(t, r["DELETE:/users/{id}"])
}

func TestFailRegister(t *testing.T) {
	dummy := func(w http.ResponseWriter, r *http.Request) {}
	h := failingHandler{Post: dummy}
	r := make(router)

	assert.Panics(t, func() {
		UsingRouter(r, h)
	})
}

func (r router) Delete(pattern string, h http.HandlerFunc) {
	r["DELETE:"+pattern] = true
}

func (r router) Get(pattern string, h http.HandlerFunc) {
	r["GET:"+pattern] = true
}

func (r router) Head(pattern string, h http.HandlerFunc) {
	r["HEAD:"+pattern] = true
}

func (r router) Options(pattern string, h http.HandlerFunc) {
	r["OPTIONS:"+pattern] = true
}

func (r router) Patch(pattern string, h http.HandlerFunc) {
	r["PATCH:"+pattern] = true
}

func (r router) Post(pattern string, h http.HandlerFunc) {
	r["POST:"+pattern] = true
}

func (r router) Put(pattern string, h http.HandlerFunc) {
	r["PUT:"+pattern] = true
}
