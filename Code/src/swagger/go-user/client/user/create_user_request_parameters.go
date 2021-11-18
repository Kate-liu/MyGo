// Code generated by go-swagger; DO NOT EDIT.

package user

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"Code/src/swagger/go-user/models"
)

// NewCreateUserRequestParams creates a new CreateUserRequestParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewCreateUserRequestParams() *CreateUserRequestParams {
	return &CreateUserRequestParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewCreateUserRequestParamsWithTimeout creates a new CreateUserRequestParams object
// with the ability to set a timeout on a request.
func NewCreateUserRequestParamsWithTimeout(timeout time.Duration) *CreateUserRequestParams {
	return &CreateUserRequestParams{
		timeout: timeout,
	}
}

// NewCreateUserRequestParamsWithContext creates a new CreateUserRequestParams object
// with the ability to set a context for a request.
func NewCreateUserRequestParamsWithContext(ctx context.Context) *CreateUserRequestParams {
	return &CreateUserRequestParams{
		Context: ctx,
	}
}

// NewCreateUserRequestParamsWithHTTPClient creates a new CreateUserRequestParams object
// with the ability to set a custom HTTPClient for a request.
func NewCreateUserRequestParamsWithHTTPClient(client *http.Client) *CreateUserRequestParams {
	return &CreateUserRequestParams{
		HTTPClient: client,
	}
}

/* CreateUserRequestParams contains all the parameters to send to the API endpoint
   for the create user request operation.

   Typically these are written to a http.Request.
*/
type CreateUserRequestParams struct {

	/* Body.

	   This text will appear as description of your request body.
	*/
	Body *models.User

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the create user request params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *CreateUserRequestParams) WithDefaults() *CreateUserRequestParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the create user request params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *CreateUserRequestParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the create user request params
func (o *CreateUserRequestParams) WithTimeout(timeout time.Duration) *CreateUserRequestParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the create user request params
func (o *CreateUserRequestParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the create user request params
func (o *CreateUserRequestParams) WithContext(ctx context.Context) *CreateUserRequestParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the create user request params
func (o *CreateUserRequestParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the create user request params
func (o *CreateUserRequestParams) WithHTTPClient(client *http.Client) *CreateUserRequestParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the create user request params
func (o *CreateUserRequestParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the create user request params
func (o *CreateUserRequestParams) WithBody(body *models.User) *CreateUserRequestParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the create user request params
func (o *CreateUserRequestParams) SetBody(body *models.User) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *CreateUserRequestParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error
	if o.Body != nil {
		if err := r.SetBodyParam(o.Body); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
