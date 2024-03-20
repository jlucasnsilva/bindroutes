package bindroutes

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type (
	testRouter map[string]bool

	testHandler struct {
		BasePath `handle:"/users"`

		Post    http.HandlerFunc `handle:"POST /"`
		Get     http.HandlerFunc `handle:"GET /{id}"`
		Put     http.HandlerFunc `handle:"PUT /{id}"`
		Delete  http.HandlerFunc `handle:"DELETE /{id}"`
		Ignored string
	}

	testGroupHandler struct {
		BasePath `handle:"/users"`

		Post   http.HandlerFunc `handle:"POST /,group=group_a"`
		Get    http.HandlerFunc `handle:"GET /{id},group=group_b"`
		Put    http.HandlerFunc `handle:"PUT /{id},group=group_a"`
		Delete http.HandlerFunc `handle:"DELETE /{id},group=group_c"`
	}

	failingHandler struct {
		Post http.HandlerFunc `handle:"Get,/users"`
	}
)

func TestSplitTag(t *testing.T) {
	assert.Panics(
		t,
		func() {
			splitTag("GET")
		},
		"should panic with incomplete string",
	)

	assert.Panics(
		t,
		func() {
			splitTag("NINJA /")
		},
		"should panic with wrong method",
	)

	method, pattern, group := splitTag("POST /,group=ninja")
	assert.Equal(t, "POST", method)
	assert.Equal(t, "/", pattern)
	assert.Equal(t, "ninja", group)
}

func TestUsingRouter(t *testing.T) {
	h := testHandler{
		Post:   dummyHandler,
		Get:    dummyHandler,
		Put:    dummyHandler,
		Delete: dummyHandler,
	}
	r := make(testRouter)

	UsingRouter(r, &h)

	assert.True(t, r["POST /users"])
	assert.True(t, r["GET /users/{id}"])
	assert.True(t, r["PUT /users/{id}"])
	assert.True(t, r["DELETE /users/{id}"])
}

func TestUsingRouters(t *testing.T) {
	h := testGroupHandler{
		Post:   dummyHandler,
		Get:    dummyHandler,
		Put:    dummyHandler,
		Delete: dummyHandler,
	}
	rs := map[string]Router{
		"group_a": make(testRouter),
		"group_b": make(testRouter),
		"group_c": make(testRouter),
	}

	UsingRouters(rs, &h)

	ra := rs["group_a"].(testRouter)
	rb := rs["group_b"].(testRouter)
	rc := rs["group_c"].(testRouter)

	assert.True(t, ra["POST /users"])
	assert.True(t, rb["GET /users/{id}"])
	assert.True(t, ra["PUT /users/{id}"])
	assert.True(t, rc["DELETE /users/{id}"])
}

func TestFailRegister(t *testing.T) {
	h := failingHandler{Post: dummyHandler}
	r := make(testRouter)

	assert.Panics(t, func() {
		UsingRouter(r, h)
	})
}

func TestGroupHandlerFuncs(t *testing.T) {
	h := testGroupHandler{
		Post:   dummyHandler,
		Get:    dummyHandler,
		Put:    dummyHandler,
		Delete: dummyHandler,
	}

	hg := groupHandlerFuncs([]any{&h})
	assert.Equal(t, 2, len(hg["group_a"]))
	assert.Equal(t, 1, len(hg["group_b"]))
	assert.Equal(t, 1, len(hg["group_c"]))
	assert.Equal(t, 0, len(hg["group_d"]))
}

func (r testRouter) Delete(pattern string, h http.HandlerFunc) {
	r["DELETE "+pattern] = true
}

func (r testRouter) Get(pattern string, h http.HandlerFunc) {
	r["GET "+pattern] = true
}

func (r testRouter) Head(pattern string, h http.HandlerFunc) {
	r["HEAD "+pattern] = true
}

func (r testRouter) Options(pattern string, h http.HandlerFunc) {
	r["OPTIONS "+pattern] = true
}

func (r testRouter) Patch(pattern string, h http.HandlerFunc) {
	r["PATCH "+pattern] = true
}

func (r testRouter) Post(pattern string, h http.HandlerFunc) {
	r["POST "+pattern] = true
}

func (r testRouter) Put(pattern string, h http.HandlerFunc) {
	r["PUT "+pattern] = true
}

func dummyHandler(w http.ResponseWriter, r *http.Request) {
	// nothing ...
}
