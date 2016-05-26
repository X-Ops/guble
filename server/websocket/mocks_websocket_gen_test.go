// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/smancke/guble/server/websocket (interfaces: WSConnection)

package websocket

import (
	gomock "github.com/golang/mock/gomock"
)

// Mock of WSConnection interface
type MockWSConnection struct {
	ctrl     *gomock.Controller
	recorder *_MockWSConnectionRecorder
}

// Recorder for MockWSConnection (not exported)
type _MockWSConnectionRecorder struct {
	mock *MockWSConnection
}

func NewMockWSConnection(ctrl *gomock.Controller) *MockWSConnection {
	mock := &MockWSConnection{ctrl: ctrl}
	mock.recorder = &_MockWSConnectionRecorder{mock}
	return mock
}

func (_m *MockWSConnection) EXPECT() *_MockWSConnectionRecorder {
	return _m.recorder
}

func (_m *MockWSConnection) Close() {
	_m.ctrl.Call(_m, "Close")
}

func (_mr *_MockWSConnectionRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Close")
}

func (_m *MockWSConnection) Receive(_param0 *[]byte) error {
	ret := _m.ctrl.Call(_m, "Receive", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockWSConnectionRecorder) Receive(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Receive", arg0)
}

func (_m *MockWSConnection) Send(_param0 []byte) error {
	ret := _m.ctrl.Call(_m, "Send", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockWSConnectionRecorder) Send(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Send", arg0)
}