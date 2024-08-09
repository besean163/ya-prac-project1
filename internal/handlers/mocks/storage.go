// Code generated by MockGen. DO NOT EDIT.
// Source: internal/handlers/handlers.go

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// GetRows mocks base method.
func (m *MockStorage) GetRows() []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRows")
	ret0, _ := ret[0].([]string)
	return ret0
}

// GetRows indicates an expected call of GetRows.
func (mr *MockStorageMockRecorder) GetRows() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRows", reflect.TypeOf((*MockStorage)(nil).GetRows))
}

// GetValue mocks base method.
func (m *MockStorage) GetValue(t, name string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValue", t, name)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetValue indicates an expected call of GetValue.
func (mr *MockStorageMockRecorder) GetValue(t, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValue", reflect.TypeOf((*MockStorage)(nil).GetValue), t, name)
}

// SetValue mocks base method.
func (m *MockStorage) SetValue(t, name, value string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetValue", t, name, value)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetValue indicates an expected call of SetValue.
func (mr *MockStorageMockRecorder) SetValue(t, name, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetValue", reflect.TypeOf((*MockStorage)(nil).SetValue), t, name, value)
}
