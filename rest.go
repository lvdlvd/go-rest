/*
Package rest provides a structure to
register handlers in a REST oriented URL space.

Example:


*/
package rest

import "net/http"

// Permission represents a set of permissions for each of the 5 operations in a handler.
type Permission uint

const (
	List Permission = 1 << iota
	Post
	Get
	Put
	Del

	Read  = List | Get
	Write = Post | Put | Del
	All   = Read | Write
)

// Everyone is an Auth function that will allow everyone the permissions p.
func Everyone(p Permission) func(r *http.Request) Permission {
	return func(r *http.Request) Permission { return p }
}

// Or logically composes permission function.  The resulting permission function
// grants a permission if any of the listed functions grant it.
func Any(f ...func(r *http.Request) Permission) func(r *http.Request) Permission {
	return func(r *http.Request) Permission {
		var p Permission
		for _, v := range f {
			p |= v(r)
		}
		return p
	}
}

// A Handler bundles related methods to be registered on a path.
//
// Auth is function that should inspect the requests credentials and return the permissions for this handlers' methods.
// The user credentials are typically stored in the request, (eg http://www.gorillatoolkit.org/pkg/context),
// the resource is the request's URL.Path, and the action is the url.Method.  The Auth function should
// return the set of allowed actions given the resource and the credentials.
//
// If Auth is nil, everyone can read, no-one can post/put/delete.
//
// A Handler with List and Post should be registered on a collection path, eg "/users"
// A Handler with Get, Put and Delete should be registered on an item path, et "/users/{id}"
// If both List and Get are defined, List is ignored and Get is used
type Handler struct {
	Auth func(r *http.Request) Permission //  Should answer the question: is (resource, user, action) permitted.

	List, // List all elements on a collection.
	Post, // Create a new element, should typically return the created id or the whole element.
	Get, // Retrieve an element by its id.
	Put, // Put an element at a given id, or replace parts of it, should typically return the updated element.
	Del http.Handler // Delete an element at a given id.
}

// If there is no handler for the request method, returns 'MethodNotAllowed'
// otherwise, calls the Auth function, and if the corresponding permission is missing
// returns 'Forbidden'.  Otherwise calls the registered handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hnd, want := h.handler(r.Method)
	if hnd == nil {
		w.Header()["Allow"] = h.allowed()
		http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
		return
	}
	got := Read
	if h.Auth != nil {
		got = h.Auth(r)
	}
	if want&got != want {
		http.Error(w, "Permission denied.", http.StatusForbidden)
		return
	}
	hnd.ServeHTTP(w, r)
}

func (h *Handler) handler(method string) (http.Handler, Permission) {
	switch method {
	case "GET":
		if h.Get != nil {
			return h.Get, Get
		}
		return h.List, List
	case "POST":
		return h.Post, Post
	case "PUT":
		return h.Put, Put
	case "DELETE":
		return h.Del, Del
	}
	return nil, 0
}

func (h *Handler) allowed() []string {
	v := make([]string, 0, 3)
	if h.List != nil || h.Get != nil {
		v = append(v, "GET")
	}
	if h.Post != nil {
		v = append(v, "POST")
	}
	if h.Put != nil {
		v = append(v, "PUT")
	}
	if h.Del != nil {
		v = append(v, "DELETE")
	}
	return v
}
