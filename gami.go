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

	done chan struct{}

	response map[string]chan *AMIResponse

	// Events for client parse
	Events chan *AMIEvent

	// Error Raise on logic
	Error chan error

	//NetError a network error
	NetError chan error
}

// AMIResponse from action
type AMIResponse struct {
	ID     string
	Status string
	Params map[string]string
}

// AMIEvent it's a representation of Event readed
type AMIEvent struct {
	//Identification of event Event: xxxx
	ID string

	Privilege []string

	// Params  of arguments received
	Params map[string]string
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
		client.response[params["ActionID"]] = make(chan *AMIResponse, 1)
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
	fmt.Println("Rec01")
	select {
	case response := <-resp:
		fmt.Println("Rec02")
		return response, nil
	case <-client.done:
		fmt.Println("Rec03")
		return nil, errors.New("may not send action on closed pipeline")
	}
}

func handleConnection(conn *textproto.Conn, done <-chan struct{}) (chan textproto.MIMEHeader, chan error) {
	datap := make(chan textproto.MIMEHeader)
	errp := make(chan error)
	go func() {
		defer func() {
			close(datap)
			close(errp)
		}()
		for {
			select {
			case <-done:
				return
			default:
				if data, err := conn.ReadMIMEHeader(); err != nil {
					errp <- err
					return
				} else {
					datap <- data
				}
			}
		}
	}()
	return datap, errp
}

// Run process socket waiting events and responses
func (client *AMIClient) run() {
	datap, errp := handleConnection(client.conn, client.done)
	go func() {
		for err := range errp {
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
	go func() {
		for data := range datap {
			if data.Get("Response") != "" {
				if response, err := newResponse(&data); err == nil {
					client.response[response.ID] <- response
					close(client.response[response.ID])
					delete(client.response, response.ID)
				} else {
					client.Error <- err
				}
			}
			if data.Get("Event") != "" {
				if event, err := newEvent(&data); err == nil {
					client.Events <- event
				} else {
					client.Error <- err
				}
			}
		}
	}()
}

//newResponse build a response for action
func newResponse(data *textproto.MIMEHeader) (*AMIResponse, error) {
	if data.Get("Response") == "" {
		return nil, errors.New("Not Response")
	}
	response := &AMIResponse{"", "", make(map[string]string)}
	for k, v := range *data {
		if k == "Response" {
			continue
		}
		if k == "Actionid" {
			continue
		}
		response.Params[k] = v[0]
	}
	response.ID = data.Get("Actionid")
	response.Status = data.Get("Response")
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
		done:        make(chan struct{}),
		response:    make(map[string]chan *AMIResponse),
		Events:      make(chan *AMIEvent, 100),
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
func (client *AMIClient) Close() {
	client.Action("Logoff", nil)
	client.conn.Close()
	close(client.done)
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
