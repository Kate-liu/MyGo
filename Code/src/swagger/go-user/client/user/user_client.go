// Code generated by go-swagger; DO NOT EDIT.

package user

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// New creates a new user API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) ClientService {
	return &Client{transport: transport, formats: formats}
}

/*
Client for user API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

// ClientOption is the option for Client methods
type ClientOption func(*runtime.ClientOperation)

// ClientService is the interface for Client methods
type ClientService interface {
	CreateUserRequest(params *CreateUserRequestParams, opts ...ClientOption) (*CreateUserRequestOK, error)

	GetUserRequest(params *GetUserRequestParams, opts ...ClientOption) (*GetUserRequestOK, error)

	SetTransport(transport runtime.ClientTransport)
}

/*
  CreateUserRequest creates a user in memory
*/
func (a *Client) CreateUserRequest(params *CreateUserRequestParams, opts ...ClientOption) (*CreateUserRequestOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCreateUserRequestParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "createUserRequest",
		Method:             "POST",
		PathPattern:        "/users",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &CreateUserRequestReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CreateUserRequestOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*CreateUserRequestDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetUserRequest gets a user from memory
*/
func (a *Client) GetUserRequest(params *GetUserRequestParams, opts ...ClientOption) (*GetUserRequestOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetUserRequestParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "getUserRequest",
		Method:             "GET",
		PathPattern:        "/users/{name}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetUserRequestReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetUserRequestOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetUserRequestDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
