//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2025 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//

// Code generated by go-swagger; DO NOT EDIT.

package objects

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

	"github.com/weaviate/weaviate/entities/models"
)

// NewObjectsUpdateParams creates a new ObjectsUpdateParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewObjectsUpdateParams() *ObjectsUpdateParams {
	return &ObjectsUpdateParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewObjectsUpdateParamsWithTimeout creates a new ObjectsUpdateParams object
// with the ability to set a timeout on a request.
func NewObjectsUpdateParamsWithTimeout(timeout time.Duration) *ObjectsUpdateParams {
	return &ObjectsUpdateParams{
		timeout: timeout,
	}
}

// NewObjectsUpdateParamsWithContext creates a new ObjectsUpdateParams object
// with the ability to set a context for a request.
func NewObjectsUpdateParamsWithContext(ctx context.Context) *ObjectsUpdateParams {
	return &ObjectsUpdateParams{
		Context: ctx,
	}
}

// NewObjectsUpdateParamsWithHTTPClient creates a new ObjectsUpdateParams object
// with the ability to set a custom HTTPClient for a request.
func NewObjectsUpdateParamsWithHTTPClient(client *http.Client) *ObjectsUpdateParams {
	return &ObjectsUpdateParams{
		HTTPClient: client,
	}
}

/*
ObjectsUpdateParams contains all the parameters to send to the API endpoint

	for the objects update operation.

	Typically these are written to a http.Request.
*/
type ObjectsUpdateParams struct {

	// Body.
	Body *models.Object

	/* ConsistencyLevel.

	   Determines how many replicas must acknowledge a request before it is considered successful
	*/
	ConsistencyLevel *string

	/* ID.

	   Unique ID of the Object.

	   Format: uuid
	*/
	ID strfmt.UUID

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the objects update params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ObjectsUpdateParams) WithDefaults() *ObjectsUpdateParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the objects update params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ObjectsUpdateParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the objects update params
func (o *ObjectsUpdateParams) WithTimeout(timeout time.Duration) *ObjectsUpdateParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the objects update params
func (o *ObjectsUpdateParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the objects update params
func (o *ObjectsUpdateParams) WithContext(ctx context.Context) *ObjectsUpdateParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the objects update params
func (o *ObjectsUpdateParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the objects update params
func (o *ObjectsUpdateParams) WithHTTPClient(client *http.Client) *ObjectsUpdateParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the objects update params
func (o *ObjectsUpdateParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the objects update params
func (o *ObjectsUpdateParams) WithBody(body *models.Object) *ObjectsUpdateParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the objects update params
func (o *ObjectsUpdateParams) SetBody(body *models.Object) {
	o.Body = body
}

// WithConsistencyLevel adds the consistencyLevel to the objects update params
func (o *ObjectsUpdateParams) WithConsistencyLevel(consistencyLevel *string) *ObjectsUpdateParams {
	o.SetConsistencyLevel(consistencyLevel)
	return o
}

// SetConsistencyLevel adds the consistencyLevel to the objects update params
func (o *ObjectsUpdateParams) SetConsistencyLevel(consistencyLevel *string) {
	o.ConsistencyLevel = consistencyLevel
}

// WithID adds the id to the objects update params
func (o *ObjectsUpdateParams) WithID(id strfmt.UUID) *ObjectsUpdateParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the objects update params
func (o *ObjectsUpdateParams) SetID(id strfmt.UUID) {
	o.ID = id
}

// WriteToRequest writes these params to a swagger request
func (o *ObjectsUpdateParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error
	if o.Body != nil {
		if err := r.SetBodyParam(o.Body); err != nil {
			return err
		}
	}

	if o.ConsistencyLevel != nil {

		// query param consistency_level
		var qrConsistencyLevel string

		if o.ConsistencyLevel != nil {
			qrConsistencyLevel = *o.ConsistencyLevel
		}
		qConsistencyLevel := qrConsistencyLevel
		if qConsistencyLevel != "" {

			if err := r.SetQueryParam("consistency_level", qConsistencyLevel); err != nil {
				return err
			}
		}
	}

	// path param id
	if err := r.SetPathParam("id", o.ID.String()); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
