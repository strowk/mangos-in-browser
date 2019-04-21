// MIT Licensed

package wasm

import (
	"errors"
	"log"
	"net"

	"github.com/strowk/websocket"

	mangos "nanomsg.org/go-mangos"
	"nanomsg.org/go/mangos/v2/transport"
)

// MangosTransport is a transport which implements only
// dialer using javascript websocket, when runned
// in Web Assembly environment in browser.
type MangosTransport struct {
}

// NewWASMTransport implements WebSocket transport
// from Web Assembly environment in browser.
// Read more about Web Assembly in https://webassembly.org/
func NewTransport() mangos.Transport {
	return &MangosTransport{}
}

func init() {
	transport.RegisterTransport(MangosTransport)
}

// NewListener would return error, because browser cannot listen.
func (nng *MangosTransport) NewListener(
	addr string,
	sock mangos.Socket,
) (mangos.PipeListener, error) {
	return nil, errors.New("Cannot support listen in WASM")
}

// Scheme would be ws for WebSocket.
func (nng *MangosTransport) Scheme() string {
	return "ws"
}

// DialerWS dials websocket.
type DialerWS struct {
	sock mangos.Socket
	url  string
}

// NewDialer creates WASMDialerWST.
func (nng *MangosTransport) NewDialer(
	url string,
	sock mangos.Socket) (mangos.PipeDialer, error) {
	return &DialerWS{
		sock: sock,
		url:  url,
	}, nil
}

// PipeWS works with websocket connection.
type PipeWS struct {
	open  bool
	conn  *net.Conn
	proto mangos.Protocol
}

// Send sends a complete message to websocket.
func (pipe *PipeWS) Send(msg *mangos.Message) error {
	if msg.Expired() {
		msg.Free()
		return nil
	}
	var buf []byte
	if len(msg.Header) > 0 {
		buf = make([]byte, 0, len(msg.Header)+len(msg.Body))
		buf = append(buf, msg.Header...)
		buf = append(buf, msg.Body...)
	} else {
		buf = msg.Body
	}

	_, err := (*pipe.conn).Write(buf)
	if err != nil {
		return err
	}
	msg.Free()
	return nil
}

// Recv receives a complete message from websocket.
func (pipe *PipeWS) Recv() (*mangos.Message, error) {
	buf := make([]byte, 1024*1024)
	n, err := (*pipe.conn).Read(buf)
	if err != nil {
		return nil, err
	}

	msg := mangos.NewMessage(0)
	msg.Body = buf[:n]
	return msg, nil
}

// Close closes the websocket.
func (pipe *PipeWS) Close() error {
	pipe.open = false
	return (*pipe.conn).Close()
}

// LocalProtocol returns the 16-bit SP protocol number used by the
// local side.
func (pipe *PipeWS) LocalProtocol() uint16 {
	return pipe.proto.Number()
}

// RemoteProtocol returns the 16-bit SP protocol number used by the
// remote side. This will normally be received from the peer during
// connection establishment.
func (pipe *PipeWS) RemoteProtocol() uint16 {
	return pipe.proto.PeerNumber()
}

// IsOpen returns true if the underlying connection is open.
func (pipe *PipeWS) IsOpen() bool {
	return pipe.open
}

// GetProp returns an arbitrary transport specific property.
// These are like options, but are read-only and specific to a single
// connection. If the property doesn't exist, then ErrBadProperty
// should be returned.
func (pipe *PipeWS) GetProp(string) (interface{}, error) {
	return nil, mangos.ErrBadProperty
}

// Dial is used to initiate a connection to a remote peer.
// It would open websocket connection to url specified in
// dialer.
func (dialer *DialerWS) Dial() (mangos.Pipe, error) {
	conn, err := websocket.DialWithSubprotocols(
		dialer.url,
		[]string{
			dialer.sock.GetProtocol().PeerName() + ".sp.nanomsg.org",
		},
	) // Blocks until connection is established.
	if err != nil {
		log.Printf("Failed to establish websocket")
		return nil, err
	}
	return &PipeWS{
		proto: dialer.sock.GetProtocol(),
		conn:  &conn,
		open:  true,
	}, nil
}

// SetOption sets a local option on the dialer.
// ErrBadOption can be returned for unrecognized options.
// ErrBadValue can be returned for incorrect value types.
func (dialer *DialerWS) SetOption(
	name string,
	value interface{}) error {
	return mangos.ErrBadOption
}

// GetOption gets a local option from the dialer.
// ErrBadOption can be returned for unrecognized options.
func (dialer *DialerWS) GetOption(name string) (
	value interface{},
	err error) {
	return nil, mangos.ErrBadOption
}
