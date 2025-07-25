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

// Code generated by mockery v2.53.2. DO NOT EDIT.

package compressionhelpers

import mock "github.com/stretchr/testify/mock"

// MockCompressionStats is an autogenerated mock type for the CompressionStats type
type MockCompressionStats struct {
	mock.Mock
}

type MockCompressionStats_Expecter struct {
	mock *mock.Mock
}

func (_m *MockCompressionStats) EXPECT() *MockCompressionStats_Expecter {
	return &MockCompressionStats_Expecter{mock: &_m.Mock}
}

// CompressionRatio provides a mock function with given fields: dimensions
func (_m *MockCompressionStats) CompressionRatio(dimensions int) float64 {
	ret := _m.Called(dimensions)

	if len(ret) == 0 {
		panic("no return value specified for CompressionRatio")
	}

	var r0 float64
	if rf, ok := ret.Get(0).(func(int) float64); ok {
		r0 = rf(dimensions)
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}

// MockCompressionStats_CompressionRatio_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompressionRatio'
type MockCompressionStats_CompressionRatio_Call struct {
	*mock.Call
}

// CompressionRatio is a helper method to define mock.On call
//   - dimensions int
func (_e *MockCompressionStats_Expecter) CompressionRatio(dimensions interface{}) *MockCompressionStats_CompressionRatio_Call {
	return &MockCompressionStats_CompressionRatio_Call{Call: _e.mock.On("CompressionRatio", dimensions)}
}

func (_c *MockCompressionStats_CompressionRatio_Call) Run(run func(dimensions int)) *MockCompressionStats_CompressionRatio_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int))
	})
	return _c
}

func (_c *MockCompressionStats_CompressionRatio_Call) Return(_a0 float64) *MockCompressionStats_CompressionRatio_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCompressionStats_CompressionRatio_Call) RunAndReturn(run func(int) float64) *MockCompressionStats_CompressionRatio_Call {
	_c.Call.Return(run)
	return _c
}

// CompressionType provides a mock function with no fields
func (_m *MockCompressionStats) CompressionType() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for CompressionType")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockCompressionStats_CompressionType_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompressionType'
type MockCompressionStats_CompressionType_Call struct {
	*mock.Call
}

// CompressionType is a helper method to define mock.On call
func (_e *MockCompressionStats_Expecter) CompressionType() *MockCompressionStats_CompressionType_Call {
	return &MockCompressionStats_CompressionType_Call{Call: _e.mock.On("CompressionType")}
}

func (_c *MockCompressionStats_CompressionType_Call) Run(run func()) *MockCompressionStats_CompressionType_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockCompressionStats_CompressionType_Call) Return(_a0 string) *MockCompressionStats_CompressionType_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCompressionStats_CompressionType_Call) RunAndReturn(run func() string) *MockCompressionStats_CompressionType_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockCompressionStats creates a new instance of MockCompressionStats. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockCompressionStats(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCompressionStats {
	mock := &MockCompressionStats{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
