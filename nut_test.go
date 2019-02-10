package nut

import (
	"net"
	"testing"
	"time"
)

type MockConn struct{}

func (c *MockConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (c *MockConn) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (c *MockConn) Close() error {
	return nil
}

func (c *MockConn) LocalAddr() net.Addr {
	return &net.IPAddr{}
}

func (c *MockConn) RemoteAddr() net.Addr {
	return &net.IPAddr{}
}

func (c *MockConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestUPSClient_AllVars(t *testing.T) {
	c := &UPSClient{}

	results, err := c.parse(`
VAR eaton test.1 "12 3"
`)
	if err != nil {
		t.Error(err)
	}

	if results["test.1"] != "12 3" {
		t.Errorf("does not match expected %v", results["test.1"])
	}
}
