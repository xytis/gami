package gami

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net/textproto"
	"testing"
)

type MockMIMEConn struct {
	h chan textproto.MIMEHeader
	e chan error
}

func (mock *MockMIMEConn) PrintfLine(format string, args ...interface{}) error {
	return nil
}

func (mock *MockMIMEConn) ReadMIMEHeader() (textproto.MIMEHeader, error) {
	select {
	case h := <-mock.h:
		return h, nil
	case e := <-mock.e:
		return nil, e
	}
}

func (mock *MockMIMEConn) Close() error {
	return nil
}

func TestPoll(t *testing.T) {
	client := AMIClient{}
	mock := MockMIMEConn{
		h: make(chan textproto.MIMEHeader),
		e: make(chan error),
	}
	client.conn = &mock
	client.raw = make(chan textproto.MIMEHeader)
	client.closing = make(chan chan error)
	client.Fatal = make(chan error)

	go func() {
		//Test data receiving
		data := textproto.MIMEHeader{}
		data.Set("A", "A")
		mock.h <- data

		assert.Equal(t, data, <-client.raw)

		//Test closing
		mock.e <- errors.New("end of test")
		<-client.Fatal
	}()

	client.poll()
}

func TestMain(t *testing.T) {
	client := AMIClient{}
	mock := MockMIMEConn{}

	client.conn = &mock
	client.raw = make(chan textproto.MIMEHeader)
	client.closing = make(chan chan error)
	client.Events = make(chan *AMIEvent)
	client.Errors = make(chan error)

	go func() {

		event := textproto.MIMEHeader{}
		event.Set("Event", "TestEvent")
		client.raw <- event
		assert.Equal(t, "TestEvent", (<-client.Events).ID)

		//Test closing
		errc := make(chan error)
		client.closing <- errc
		<-errc
	}()

	client.main()
}
