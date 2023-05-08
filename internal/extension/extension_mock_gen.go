// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stateful/runme/internal/extension (interfaces: Extensioner)

// Package extension is a generated GoMock package.
package extension

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockExtensioner is a mock of Extensioner interface.
type MockExtensioner struct {
	ctrl     *gomock.Controller
	recorder *MockExtensionerMockRecorder
}

// MockExtensionerMockRecorder is the mock recorder for MockExtensioner.
type MockExtensionerMockRecorder struct {
	mock *MockExtensioner
}

// NewMockExtensioner creates a new mock instance.
func NewMockExtensioner(ctrl *gomock.Controller) *MockExtensioner {
	mock := &MockExtensioner{ctrl: ctrl}
	mock.recorder = &MockExtensionerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExtensioner) EXPECT() *MockExtensionerMockRecorder {
	return m.recorder
}

// Install mocks base method.
func (m *MockExtensioner) Install() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Install")
	ret0, _ := ret[0].(error)
	return ret0
}

// Install indicates an expected call of Install.
func (mr *MockExtensionerMockRecorder) Install() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Install", reflect.TypeOf((*MockExtensioner)(nil).Install))
}

// IsInstalled mocks base method.
func (m *MockExtensioner) IsInstalled() (string, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsInstalled")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// IsInstalled indicates an expected call of IsInstalled.
func (mr *MockExtensionerMockRecorder) IsInstalled() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsInstalled", reflect.TypeOf((*MockExtensioner)(nil).IsInstalled))
}

// Update mocks base method.
func (m *MockExtensioner) Update() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update")
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockExtensionerMockRecorder) Update() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockExtensioner)(nil).Update))
}
