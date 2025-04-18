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

// Code generated by mockery v2.53.2. DO NOT EDIT.

package authorization

import (
	mock "github.com/stretchr/testify/mock"
	models "github.com/weaviate/weaviate/entities/models"
)

// MockAuthorizer is an autogenerated mock type for the Authorizer type
type MockAuthorizer struct {
	mock.Mock
}

type MockAuthorizer_Expecter struct {
	mock *mock.Mock
}

func (_m *MockAuthorizer) EXPECT() *MockAuthorizer_Expecter {
	return &MockAuthorizer_Expecter{mock: &_m.Mock}
}

// Authorize provides a mock function with given fields: principal, verb, resources
func (_m *MockAuthorizer) Authorize(principal *models.Principal, verb string, resources ...string) error {
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

// MockAuthorizer_Authorize_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Authorize'
type MockAuthorizer_Authorize_Call struct {
	*mock.Call
}

// Authorize is a helper method to define mock.On call
//   - principal *models.Principal
//   - verb string
//   - resources ...string
func (_e *MockAuthorizer_Expecter) Authorize(principal interface{}, verb interface{}, resources ...interface{}) *MockAuthorizer_Authorize_Call {
	return &MockAuthorizer_Authorize_Call{Call: _e.mock.On("Authorize",
		append([]interface{}{principal, verb}, resources...)...)}
}

func (_c *MockAuthorizer_Authorize_Call) Run(run func(principal *models.Principal, verb string, resources ...string)) *MockAuthorizer_Authorize_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(args[0].(*models.Principal), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *MockAuthorizer_Authorize_Call) Return(_a0 error) *MockAuthorizer_Authorize_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockAuthorizer_Authorize_Call) RunAndReturn(run func(*models.Principal, string, ...string) error) *MockAuthorizer_Authorize_Call {
	_c.Call.Return(run)
	return _c
}

// AuthorizeSilent provides a mock function with given fields: principal, verb, resources
func (_m *MockAuthorizer) AuthorizeSilent(principal *models.Principal, verb string, resources ...string) error {
	_va := make([]interface{}, len(resources))
	for _i := range resources {
		_va[_i] = resources[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, principal, verb)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for AuthorizeSilent")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*models.Principal, string, ...string) error); ok {
		r0 = rf(principal, verb, resources...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockAuthorizer_AuthorizeSilent_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AuthorizeSilent'
type MockAuthorizer_AuthorizeSilent_Call struct {
	*mock.Call
}

// AuthorizeSilent is a helper method to define mock.On call
//   - principal *models.Principal
//   - verb string
//   - resources ...string
func (_e *MockAuthorizer_Expecter) AuthorizeSilent(principal interface{}, verb interface{}, resources ...interface{}) *MockAuthorizer_AuthorizeSilent_Call {
	return &MockAuthorizer_AuthorizeSilent_Call{Call: _e.mock.On("AuthorizeSilent",
		append([]interface{}{principal, verb}, resources...)...)}
}

func (_c *MockAuthorizer_AuthorizeSilent_Call) Run(run func(principal *models.Principal, verb string, resources ...string)) *MockAuthorizer_AuthorizeSilent_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(args[0].(*models.Principal), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *MockAuthorizer_AuthorizeSilent_Call) Return(_a0 error) *MockAuthorizer_AuthorizeSilent_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockAuthorizer_AuthorizeSilent_Call) RunAndReturn(run func(*models.Principal, string, ...string) error) *MockAuthorizer_AuthorizeSilent_Call {
	_c.Call.Return(run)
	return _c
}

// FilterAuthorizedResources provides a mock function with given fields: principal, verb, resources
func (_m *MockAuthorizer) FilterAuthorizedResources(principal *models.Principal, verb string, resources ...string) ([]string, error) {
	_va := make([]interface{}, len(resources))
	for _i := range resources {
		_va[_i] = resources[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, principal, verb)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for FilterAuthorizedResources")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(*models.Principal, string, ...string) ([]string, error)); ok {
		return rf(principal, verb, resources...)
	}
	if rf, ok := ret.Get(0).(func(*models.Principal, string, ...string) []string); ok {
		r0 = rf(principal, verb, resources...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(*models.Principal, string, ...string) error); ok {
		r1 = rf(principal, verb, resources...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockAuthorizer_FilterAuthorizedResources_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FilterAuthorizedResources'
type MockAuthorizer_FilterAuthorizedResources_Call struct {
	*mock.Call
}

// FilterAuthorizedResources is a helper method to define mock.On call
//   - principal *models.Principal
//   - verb string
//   - resources ...string
func (_e *MockAuthorizer_Expecter) FilterAuthorizedResources(principal interface{}, verb interface{}, resources ...interface{}) *MockAuthorizer_FilterAuthorizedResources_Call {
	return &MockAuthorizer_FilterAuthorizedResources_Call{Call: _e.mock.On("FilterAuthorizedResources",
		append([]interface{}{principal, verb}, resources...)...)}
}

func (_c *MockAuthorizer_FilterAuthorizedResources_Call) Run(run func(principal *models.Principal, verb string, resources ...string)) *MockAuthorizer_FilterAuthorizedResources_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(args[0].(*models.Principal), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *MockAuthorizer_FilterAuthorizedResources_Call) Return(_a0 []string, _a1 error) *MockAuthorizer_FilterAuthorizedResources_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockAuthorizer_FilterAuthorizedResources_Call) RunAndReturn(run func(*models.Principal, string, ...string) ([]string, error)) *MockAuthorizer_FilterAuthorizedResources_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockAuthorizer creates a new instance of MockAuthorizer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockAuthorizer(t interface {
	mock.TestingT
	Cleanup(func())
},
) *MockAuthorizer {
	mock := &MockAuthorizer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
