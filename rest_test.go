package rest_test

import (
	rest "."

	"net/http"
	"testing"

	"github.com/gorilla/mux"
)

var newOrder, listOrders, getOrder, putOrder, delOrder http.Handler

func MembersCanWrite(r *http.Request) rest.Permission {
	if true /* r, credentials contain membership */ {
		return rest.Write
	}
	return rest.Read
}

func TestThatItCompiles(t *testing.T) {

	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()

	api.Path("/orders").Handler(&rest.Handler{
		Auth: MembersCanWrite,
		List: listOrders,
		Put:  newOrder,
	})

	api.Path("/orders/{id}").Handler(&rest.Handler{
		Auth: MembersCanWrite,
		Get:  getOrder,
		Put:  putOrder,
		Del:  delOrder,
	})

}
