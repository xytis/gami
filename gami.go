package gami

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Raise when not response expected protocol AMI
var ErrNotAMI = errors.New("Server not AMI interface")

// Params for the actions
type Params map[string]string

// AMIClient a connection to AMI server
type AMIClient struct {
	conn *textproto.Conn

	address     string
	amiUser     string
	amiPass     string
	amiAuth     bool
	useTLS      bool
	unsecureTLS bool
	opNumber    int
	opPrefix    string

	raw      chan textproto.MIMEHeader
	closing  chan chan error
	events   chan *AMIEvent
	errors   chan error
	response map[string]chan *AMIResponse

	Error    chan error
	NetError chan error
}

// AMIResponse from action
type AMIResponse struct {
	ID     string
	Status string
	Params Params
}

// AMIAction
type AMIAction struct {
	Response chan AMIResponse
	Error    chan error
}

// AMIEvent it's a representation of Event readed
type AMIEvent struct {
	ID        string
	Privilege []string
	Params    Params
}

// AsyncAction returns chan for wait response of action with parameter *ActionID* this can be helpful for
// massive actions,
func (client *AMIClient) AsyncAction(action string, params Params) (<-chan *AMIResponse, error) {
	if params == nil {
		params = Params{}
	}

	if _, ok := params["ActionID"]; !ok {
		params["ActionID"] = client.opPrefix + strconv.Itoa(client.opNumber)
		client.opNumber += 1
	}

	if err := client.conn.PrintfLine("Action: %s", strings.TrimSpace(action)); err != nil {
		return nil, err
	}

	if _, ok := client.response[params["ActionID"]]; !ok {
		client.response[params["ActionID"]] = make(chan *AMIResponse)
	}

	for k, v := range params {
		if err := client.conn.PrintfLine("%s: %s", k, strings.TrimSpace(v)); err != nil {
			return nil, err
		}
	}

	if err := client.conn.PrintfLine(""); err != nil {
		return nil, err
	}

	return client.response[params["ActionID"]], nil
}

// Action send with params
func (client *AMIClient) Action(action string, params Params) (*AMIResponse, error) {
	resp, err := client.AsyncAction(action, params)
	if err != nil {
		return nil, err
	}
	select {
	case response := <-resp:
		return response, nil
	case <-time.After(time.Second * 5):
		return nil, errors.New("operation timed out")
	}
}

func (client *AMIClient) handleConnection(conn *textproto.Conn) {
	client.raw = make(chan textproto.MIMEHeader)
	go func() {
		defer func() {
			close(client.raw)
		}()
		for {
			if data, err := conn.ReadMIMEHeader(); err != nil {
				client.errors <- err
				return
			} else {
				client.raw <- data
			}
		}
	}()
}

// Run process socket waiting events and responses
func (client *AMIClient) run() {

	//Chomping proc for printing runtime errors encountered
	go func() {
		for {
			select {
			case err, ok := <-client.errors:
				if !ok {
					return
				}
				fmt.Println(err)
			}
		}
	}()

	go func() {
		var pendingEvent []*AMIEvent
		var pendingError []error
		for {
			var currentEvent *AMIEvent
			var currentError error
			var events chan *AMIEvent
			var errors chan error
			if len(pendingEvent) > 0 {
				currentEvent = pendingEvent[0]
				events = client.events
			}
			if len(pendingError) > 0 {
				currentError = pendingError[0]
				errors = client.errors
			}
			select {
			case errc := <-client.closing:
				errc <- currentError
				close(client.raw)
				close(client.errors)
				close(client.events)
				return
			case data := <-client.raw:
				if data.Get("Response") != "" {
					if response, err := newResponse(&data); err == nil {
						//TODO: will block whole server on bad consume
						client.response[response.ID] <- response
						close(client.response[response.ID])
						delete(client.response, response.ID)
					} else {
						pendingError = append(pendingError, err)
					}
				}
				if data.Get("Event") != "" {
					if event, err := newEvent(&data); err == nil {
						pendingEvent = append(pendingEvent, event)
					} else {
						pendingError = append(pendingError, err)
					}
				}
			case events <- currentEvent:
				pendingEvent = pendingEvent[1:]
			case errors <- currentError:
				pendingError = pendingError[1:]
			}
		}
	}()

	client.handleConnection(client.conn)
	go func() {
		for err := range client.errors {
			fmt.Println("connection error", err)
			switch err {
			case syscall.ECONNABORTED:
				fallthrough
			case syscall.ECONNRESET:
				fallthrough
			case syscall.ECONNREFUSED:
				fallthrough
			case io.EOF:
				client.conn.Close()
				client.NetError <- err
			default:
				client.Error <- err
			}
		}
	}()
}

//newResponse build a response for action
func newResponse(data *textproto.MIMEHeader) (*AMIResponse, error) {
	if data.Get("Response") == "" {
		return nil, errors.New("Not Response")
	}
	response := &AMIResponse{
		ID:     "",
		Status: "",
		Params: make(map[string]string),
	}
	for k, v := range *data {
		//Allways operate on first value
		fv := v[0]
		switch {
		case k == "Response":
			response.Status = fv
		case k == "Actionid":
			response.ID = fv
		default:
			response.Params[k] = fv
		}
	}
	return response, nil
}

//newEvent build event
func newEvent(data *textproto.MIMEHeader) (*AMIEvent, error) {
	if data.Get("Event") == "" {
		return nil, errors.New("Not Event")
	}
	ev := &AMIEvent{data.Get("Event"), strings.Split(data.Get("Privilege"), ","), make(map[string]string)}
	for k, v := range *data {
		if k == "Event" || k == "Privilege" {
			continue
		}
		ev.Params[k] = v[0]
	}
	return ev, nil
}

// Dial create a new connection to AMI
func Dial(address string, options ...func(*AMIClient)) (client *AMIClient, err error) {
	client = &AMIClient{
		address:     address,
		amiUser:     "",
		amiPass:     "",
		opNumber:    0,
		opPrefix:    "r",
		response:    make(map[string]chan *AMIResponse),
		events:      make(chan *AMIEvent),
		Error:       make(chan error, 1),
		NetError:    make(chan error, 1),
		useTLS:      false,
		unsecureTLS: false,
	}
	for _, op := range options {
		op(client)
	}
	if client.conn, err = client.newConn(); err != nil {
		return nil, err
	}
	client.run()
	return client, nil
}

// Login authenticate to AMI
func (client *AMIClient) Login(username, password string) error {
	response, err := client.Action("Login", Params{"Username": username, "Secret": password})
	if err != nil {
		return err
	}

	if (*response).Status == "Error" {
		return errors.New((*response).Params["Message"])
	}

	client.amiUser = username
	client.amiPass = password
	client.amiAuth = true
	return nil
}

// Reconnect the session, autologin if possible
func (client *AMIClient) Reconnect() (err error) {
	client.conn.Close()
	if client.conn, err = client.newConn(); err != nil {
		return err
	}

	if client.amiAuth {
		if err = client.Login(client.amiUser, client.amiPass); err != nil {
			return err
		}
	}

	client.run()

	return nil
}

// Close the connection to AMI
func (client *AMIClient) Close() error {
	client.Action("Logoff", nil)
	errc := make(chan error)
	client.closing <- errc
	return <-errc
}

// NewConn create a new connection to AMI
func (client *AMIClient) newConn() (conn *textproto.Conn, err error) {
	var rwc io.ReadWriteCloser
	if client.useTLS {
		rwc, err = tls.Dial("tcp", client.address, &tls.Config{InsecureSkipVerify: client.unsecureTLS})
	} else {
		rwc, err = net.Dial("tcp", client.address)
	}

	if err != nil {
		return nil, err
	}

	conn = textproto.NewConn(rwc)
	label, err := conn.ReadLine()
	if err != nil {
		return nil, err
	}

	if strings.Contains(label, "Asterisk Call Manager") != true {
		return nil, ErrNotAMI
	}

	return conn, nil
}
