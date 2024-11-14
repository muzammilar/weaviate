//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2024 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//
// Code generated by mockery v2.44.1. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	models "github.com/weaviate/weaviate/entities/models"
)

// Authorizer is an autogenerated mock type for the Authorizer type
type Authorizer struct {
	mock.Mock
}

// Authorize provides a mock function with given fields: principal, verb, resources
func (_m *Authorizer) Authorize(principal *models.Principal, verb string, resources ...string) error {
	_va := make([]interface{}, len(resources))
	for _i := range resources {
		_va[_i] = resources[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, principal, verb)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Authorize")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*models.Principal, string, ...string) error); ok {
		r0 = rf(principal, verb, resources...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewAuthorizer creates a new instance of Authorizer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAuthorizer(t interface {
	mock.TestingT
	Cleanup(func())
}) *Authorizer {
	mock := &Authorizer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}