// Code generated by MockGen. DO NOT EDIT.
// Source: app.go
//
// Generated by this command:
//
//	mockgen -source=app.go -destination=mocks/app.go -package=mocksapp
//

// Package mocksapp is a generated GoMock package.
package mocksapp

import (
	context "context"
	reflect "reflect"

	dto "github.com/MukizuL/shortener/internal/dto"
	gomock "go.uber.org/mock/gomock"
)

// Mockrepo is a mock of repo interface.
type Mockrepo struct {
	ctrl     *gomock.Controller
	recorder *MockrepoMockRecorder
	isgomock struct{}
}

// MockrepoMockRecorder is the mock recorder for Mockrepo.
type MockrepoMockRecorder struct {
	mock *Mockrepo
}

// NewMockrepo creates a new mock instance.
func NewMockrepo(ctrl *gomock.Controller) *Mockrepo {
	mock := &Mockrepo{ctrl: ctrl}
	mock.recorder = &MockrepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mockrepo) EXPECT() *MockrepoMockRecorder {
	return m.recorder
}

// BatchCreateShortURL mocks base method.
func (m *Mockrepo) BatchCreateShortURL(ctx context.Context, data []dto.BatchRequest) ([]dto.BatchResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BatchCreateShortURL", ctx, data)
	ret0, _ := ret[0].([]dto.BatchResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BatchCreateShortURL indicates an expected call of BatchCreateShortURL.
func (mr *MockrepoMockRecorder) BatchCreateShortURL(ctx, data any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchCreateShortURL", reflect.TypeOf((*Mockrepo)(nil).BatchCreateShortURL), ctx, data)
}

// CreateShortURL mocks base method.
func (m *Mockrepo) CreateShortURL(ctx context.Context, fullURL string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateShortURL", ctx, fullURL)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateShortURL indicates an expected call of CreateShortURL.
func (mr *MockrepoMockRecorder) CreateShortURL(ctx, fullURL any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateShortURL", reflect.TypeOf((*Mockrepo)(nil).CreateShortURL), ctx, fullURL)
}

// GetLongURL mocks base method.
func (m *Mockrepo) GetLongURL(ctx context.Context, ID string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLongURL", ctx, ID)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLongURL indicates an expected call of GetLongURL.
func (mr *MockrepoMockRecorder) GetLongURL(ctx, ID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLongURL", reflect.TypeOf((*Mockrepo)(nil).GetLongURL), ctx, ID)
}

// OffloadStorage mocks base method.
func (m *Mockrepo) OffloadStorage(ctx context.Context, filepath string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OffloadStorage", ctx, filepath)
	ret0, _ := ret[0].(error)
	return ret0
}

// OffloadStorage indicates an expected call of OffloadStorage.
func (mr *MockrepoMockRecorder) OffloadStorage(ctx, filepath any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OffloadStorage", reflect.TypeOf((*Mockrepo)(nil).OffloadStorage), ctx, filepath)
}

// Ping mocks base method.
func (m *Mockrepo) Ping(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockrepoMockRecorder) Ping(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*Mockrepo)(nil).Ping), ctx)
}
