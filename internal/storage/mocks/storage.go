// Code generated by MockGen. DO NOT EDIT.
// Source: storage.go
//
// Generated by this command:
//
//	mockgen -source=storage.go -destination=mocks/storage.go -package=mockstorage
//

// Package mockstorage is a generated GoMock package.
package mockstorage

import (
	context "context"
	reflect "reflect"

	dto "github.com/MukizuL/shortener/internal/dto"
	gomock "go.uber.org/mock/gomock"
)

// MockRepo is a mock of Repo interface.
type MockRepo struct {
	ctrl     *gomock.Controller
	recorder *MockRepoMockRecorder
	isgomock struct{}
}

// MockRepoMockRecorder is the mock recorder for MockRepo.
type MockRepoMockRecorder struct {
	mock *MockRepo
}

// NewMockRepo creates a new mock instance.
func NewMockRepo(ctrl *gomock.Controller) *MockRepo {
	mock := &MockRepo{ctrl: ctrl}
	mock.recorder = &MockRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepo) EXPECT() *MockRepoMockRecorder {
	return m.recorder
}

// BatchCreateShortURL mocks base method.
func (m *MockRepo) BatchCreateShortURL(ctx context.Context, userID, urlBase string, data []dto.BatchRequest) ([]dto.BatchResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BatchCreateShortURL", ctx, userID, urlBase, data)
	ret0, _ := ret[0].([]dto.BatchResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BatchCreateShortURL indicates an expected call of BatchCreateShortURL.
func (mr *MockRepoMockRecorder) BatchCreateShortURL(ctx, userID, urlBase, data any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchCreateShortURL", reflect.TypeOf((*MockRepo)(nil).BatchCreateShortURL), ctx, userID, urlBase, data)
}

// CreateShortURL mocks base method.
func (m *MockRepo) CreateShortURL(ctx context.Context, userID, urlBase, fullURL string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateShortURL", ctx, userID, urlBase, fullURL)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateShortURL indicates an expected call of CreateShortURL.
func (mr *MockRepoMockRecorder) CreateShortURL(ctx, userID, urlBase, fullURL any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateShortURL", reflect.TypeOf((*MockRepo)(nil).CreateShortURL), ctx, userID, urlBase, fullURL)
}

// DeleteURLs mocks base method.
func (m *MockRepo) DeleteURLs(ctx context.Context, userID string, urls []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteURLs", ctx, userID, urls)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteURLs indicates an expected call of DeleteURLs.
func (mr *MockRepoMockRecorder) DeleteURLs(ctx, userID, urls any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteURLs", reflect.TypeOf((*MockRepo)(nil).DeleteURLs), ctx, userID, urls)
}

// GetLongURL mocks base method.
func (m *MockRepo) GetLongURL(ctx context.Context, ID string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLongURL", ctx, ID)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLongURL indicates an expected call of GetLongURL.
func (mr *MockRepoMockRecorder) GetLongURL(ctx, ID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLongURL", reflect.TypeOf((*MockRepo)(nil).GetLongURL), ctx, ID)
}

// GetUserURLs mocks base method.
func (m *MockRepo) GetUserURLs(ctx context.Context, userID string) ([]dto.URLPair, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserURLs", ctx, userID)
	ret0, _ := ret[0].([]dto.URLPair)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserURLs indicates an expected call of GetUserURLs.
func (mr *MockRepoMockRecorder) GetUserURLs(ctx, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserURLs", reflect.TypeOf((*MockRepo)(nil).GetUserURLs), ctx, userID)
}

// OffloadStorage mocks base method.
func (m *MockRepo) OffloadStorage(ctx context.Context, filepath string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OffloadStorage", ctx, filepath)
	ret0, _ := ret[0].(error)
	return ret0
}

// OffloadStorage indicates an expected call of OffloadStorage.
func (mr *MockRepoMockRecorder) OffloadStorage(ctx, filepath any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OffloadStorage", reflect.TypeOf((*MockRepo)(nil).OffloadStorage), ctx, filepath)
}

// Ping mocks base method.
func (m *MockRepo) Ping(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockRepoMockRecorder) Ping(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockRepo)(nil).Ping), ctx)
}
