// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/smancke/guble/client (interfaces: WSConnection,Client)

package client

import (
	gomock "github.com/golang/mock/gomock"
	
	guble "github.com/smancke/guble/guble"
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

func (_m *MockWSConnection) Close() error {
	ret := _m.ctrl.Call(_m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockWSConnectionRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Close")
}

func (_m *MockWSConnection) ReadMessage() (int, []byte, error) {
	ret := _m.ctrl.Call(_m, "ReadMessage")
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].([]byte)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (_mr *_MockWSConnectionRecorder) ReadMessage() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ReadMessage")
}

func (_m *MockWSConnection) WriteMessage(_param0 int, _param1 []byte) error {
	ret := _m.ctrl.Call(_m, "WriteMessage", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockWSConnectionRecorder) WriteMessage(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "WriteMessage", arg0, arg1)
}

// Mock of Client interface
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *_MockClientRecorder
}

// Recorder for MockClient (not exported)
type _MockClientRecorder struct {
	mock *MockClient
}

func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &_MockClientRecorder{mock}
	return mock
}

func (_m *MockClient) EXPECT() *_MockClientRecorder {
	return _m.recorder
}

func (_m *MockClient) Close() {
	_m.ctrl.Call(_m, "Close")
}

func (_mr *_MockClientRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Close")
}

func (_m *MockClient) Errors() chan *guble.NotificationMessage {
	ret := _m.ctrl.Call(_m, "Errors")
	ret0, _ := ret[0].(chan *guble.NotificationMessage)
	return ret0
}

func (_mr *_MockClientRecorder) Errors() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Errors")
}

func (_m *MockClient) IsConnected() bool {
	ret := _m.ctrl.Call(_m, "IsConnected")
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockClientRecorder) IsConnected() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "IsConnected")
}

func (_m *MockClient) Messages() chan *guble.Message {
	ret := _m.ctrl.Call(_m, "Messages")
	ret0, _ := ret[0].(chan *guble.Message)
	return ret0
}

func (_mr *_MockClientRecorder) Messages() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Messages")
}

func (_m *MockClient) Send(_param0 string, _param1 string, _param2 string) error {
	ret := _m.ctrl.Call(_m, "Send", _param0, _param1, _param2)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockClientRecorder) Send(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Send", arg0, arg1, arg2)
}

func (_m *MockClient) SendBytes(_param0 string, _param1 []byte, _param2 string) error {
	ret := _m.ctrl.Call(_m, "SendBytes", _param0, _param1, _param2)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockClientRecorder) SendBytes(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SendBytes", arg0, arg1, arg2)
}

func (_m *MockClient) SetWSConnectionFactory(_param0  WSConnectionFactory) {
	_m.ctrl.Call(_m, "SetWSConnectionFactory", _param0)
}

func (_mr *_MockClientRecorder) SetWSConnectionFactory(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetWSConnectionFactory", arg0)
}

func (_m *MockClient) Start() error {
	ret := _m.ctrl.Call(_m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockClientRecorder) Start() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Start")
}

func (_m *MockClient) StatusMessages() chan *guble.NotificationMessage {
	ret := _m.ctrl.Call(_m, "StatusMessages")
	ret0, _ := ret[0].(chan *guble.NotificationMessage)
	return ret0
}

func (_mr *_MockClientRecorder) StatusMessages() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "StatusMessages")
}

func (_m *MockClient) Subscribe(_param0 string) error {
	ret := _m.ctrl.Call(_m, "Subscribe", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockClientRecorder) Subscribe(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Subscribe", arg0)
}

func (_m *MockClient) Unsubscribe(_param0 string) error {
	ret := _m.ctrl.Call(_m, "Unsubscribe", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockClientRecorder) Unsubscribe(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Unsubscribe", arg0)
}

func (_m *MockClient) WriteRawMessage(_param0 []byte) error {
	ret := _m.ctrl.Call(_m, "WriteRawMessage", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockClientRecorder) WriteRawMessage(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "WriteRawMessage", arg0)
}
