package gami

import (
	"errors"
	"io"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

// Raise when not response expected protocol AMI
var ErrNotAMI = errors.New("Server not AMI interface")

// Params for the actions
type Params map[string]string

// sub-interface that is user by this lib
type MIMEReadWriteCloser interface {
	PrintfLine(format string, args ...interface{}) error
	ReadMIMEHeader() (textproto.MIMEHeader, error)
	Close() error
}

// AMIClient a connection to AMI server
type AMIClient struct {
	conn MIMEReadWriteCloser

	address string
	amiUser string
	amiPass string

	opNumber int
	opPrefix string

	raw      chan textproto.MIMEHeader
	closing  chan chan error
	response map[string]chan *AMIResponse

	Events chan *AMIEvent
	Errors chan error
	Fatal  chan error
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

func (client *AMIClient) poll() {
	for {
		if data, err := client.conn.ReadMIMEHeader(); err != nil {
			//When underlying connection is closed, reader 'sometimes' returns EOF
			client.Fatal <- err
			close(client.closing)
			return
		} else {
			client.raw <- data
		}
	}
}

func (client *AMIClient) main() {
	var pendingEvent []*AMIEvent
	var pendingError []error
	for {
		var currentEvent *AMIEvent
		var currentError error
		var events chan *AMIEvent
		var errors chan error
		if len(pendingEvent) > 0 {
			currentEvent = pendingEvent[0]
			events = client.Events
		}
		if len(pendingError) > 0 {
			currentError = pendingError[0]
			errors = client.Errors
		}
		select {
		case errc, ok := <-client.closing:
			if ok {
				err := client.conn.Close()
				errc <- err
			}
			//Closing to notify that we are offline
			close(client.Events)
			return
		case data, ok := <-client.raw:
			if !ok {
				// client.raw got closed?
				continue
			}
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
}

// Run process socket waiting events and responses
func (client *AMIClient) run() {
	go client.main()
	go client.poll()
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

// Create a new connection to AMI
func Connect(address string, user string, secret string) (client *AMIClient, err error) {
	client = &AMIClient{
		address:  address,
		amiUser:  user,
		amiPass:  secret,
		opNumber: 0,
		opPrefix: "r",

		response: make(map[string]chan *AMIResponse),
		raw:      make(chan textproto.MIMEHeader),
		closing:  make(chan chan error),

		Events: make(chan *AMIEvent),
		Errors: make(chan error),
		Fatal:  make(chan error),
	}
	if err = client.bind(); err != nil {
		return nil, err
	}
	client.run()
	if user != "" {
		return client, client.login(user, secret)
	}
	return client, nil
}

// Login authenticate to AMI
func (client *AMIClient) login(username, password string) error {
	response, err := client.Action("Login", Params{"Username": username, "Secret": password})
	if err != nil {
		return err
	}

	if (*response).Status == "Error" {
		return errors.New((*response).Params["Message"])
	}

	client.amiUser = username
	client.amiPass = password
	return nil
}

// Close the connection to AMI
func (client *AMIClient) Close() error {
	if _, err := client.Action("Logoff", nil); err != nil {
		return err
	}
	errc := make(chan error)
	client.closing <- errc
	return <-errc
}

//  create a new connection to AMI
func (client *AMIClient) bind() (err error) {
	var rwc io.ReadWriteCloser
	rwc, err = net.Dial("tcp", client.address)

	if err != nil {
		return err
	}

	conn := textproto.NewConn(rwc)
	label, err := conn.ReadLine()
	if err != nil {
		return err
	}

	if strings.Contains(label, "Asterisk Call Manager") != true {
		return ErrNotAMI
	}

	client.conn = conn
	return nil
}
