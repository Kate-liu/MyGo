// Code generated by go-swagger; DO NOT EDIT.

package user

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// GetUserRequestHandlerFunc turns a function with the right signature into a get user request handler
type GetUserRequestHandlerFunc func(GetUserRequestParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetUserRequestHandlerFunc) Handle(params GetUserRequestParams) middleware.Responder {
	return fn(params)
}

// GetUserRequestHandler interface for that can handle valid get user request params
type GetUserRequestHandler interface {
	Handle(GetUserRequestParams) middleware.Responder
}

// NewGetUserRequest creates a new http.Handler for the get user request operation
func NewGetUserRequest(ctx *middleware.Context, handler GetUserRequestHandler) *GetUserRequest {
	return &GetUserRequest{Context: ctx, Handler: handler}
}

/* GetUserRequest swagger:route GET /users/{name} user getUserRequest

Get a user from memory.

*/
type GetUserRequest struct {
	Context *middleware.Context
	Handler GetUserRequestHandler
}

func (o *GetUserRequest) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewGetUserRequestParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
