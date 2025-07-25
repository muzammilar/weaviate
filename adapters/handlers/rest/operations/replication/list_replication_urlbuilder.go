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

package replication

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"errors"
	"net/url"
	golangswaggerpaths "path"

	"github.com/go-openapi/swag"
)

// ListReplicationURL generates an URL for the list replication operation
type ListReplicationURL struct {
	Collection     *string
	IncludeHistory *bool
	Shard          *string
	TargetNode     *string

	_basePath string
	// avoid unkeyed usage
	_ struct{}
}

// WithBasePath sets the base path for this url builder, only required when it's different from the
// base path specified in the swagger spec.
// When the value of the base path is an empty string
func (o *ListReplicationURL) WithBasePath(bp string) *ListReplicationURL {
	o.SetBasePath(bp)
	return o
}

// SetBasePath sets the base path for this url builder, only required when it's different from the
// base path specified in the swagger spec.
// When the value of the base path is an empty string
func (o *ListReplicationURL) SetBasePath(bp string) {
	o._basePath = bp
}

// Build a url path and query string
func (o *ListReplicationURL) Build() (*url.URL, error) {
	var _result url.URL

	var _path = "/replication/replicate/list"

	_basePath := o._basePath
	if _basePath == "" {
		_basePath = "/v1"
	}
	_result.Path = golangswaggerpaths.Join(_basePath, _path)

	qs := make(url.Values)

	var collectionQ string
	if o.Collection != nil {
		collectionQ = *o.Collection
	}
	if collectionQ != "" {
		qs.Set("collection", collectionQ)
	}

	var includeHistoryQ string
	if o.IncludeHistory != nil {
		includeHistoryQ = swag.FormatBool(*o.IncludeHistory)
	}
	if includeHistoryQ != "" {
		qs.Set("includeHistory", includeHistoryQ)
	}

	var shardQ string
	if o.Shard != nil {
		shardQ = *o.Shard
	}
	if shardQ != "" {
		qs.Set("shard", shardQ)
	}

	var targetNodeQ string
	if o.TargetNode != nil {
		targetNodeQ = *o.TargetNode
	}
	if targetNodeQ != "" {
		qs.Set("targetNode", targetNodeQ)
	}

	_result.RawQuery = qs.Encode()

	return &_result, nil
}

// Must is a helper function to panic when the url builder returns an error
func (o *ListReplicationURL) Must(u *url.URL, err error) *url.URL {
	if err != nil {
		panic(err)
	}
	if u == nil {
		panic("url can't be nil")
	}
	return u
}

// String returns the string representation of the path with query string
func (o *ListReplicationURL) String() string {
	return o.Must(o.Build()).String()
}

// BuildFull builds a full url with scheme, host, path and query string
func (o *ListReplicationURL) BuildFull(scheme, host string) (*url.URL, error) {
	if scheme == "" {
		return nil, errors.New("scheme is required for a full url on ListReplicationURL")
	}
	if host == "" {
		return nil, errors.New("host is required for a full url on ListReplicationURL")
	}

	base, err := o.Build()
	if err != nil {
		return nil, err
	}

	base.Scheme = scheme
	base.Host = host
	return base, nil
}

// StringFull returns the string representation of a complete url
func (o *ListReplicationURL) StringFull(scheme, host string) string {
	return o.Must(o.BuildFull(scheme, host)).String()
}
