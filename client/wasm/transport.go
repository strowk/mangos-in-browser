// MIT Licensed

// +build wasm

package wasm

import (
	"errors"
	"log"
	"net"

	"github.com/strowk/websocket"
	mangos "nanomsg.org/go-mangos"
)

// WASMMangosTransport
type WASMMangosTransport struct {
}

// NewWASMTransport implements WebSocket transport
// from Web Assembly environment in browser.
// Read more about Web Assembly in https://webassembly.org/
func NewWASMTransport() mangos.Transport {
	return &WASMMangosTransport{}
}

// NewListener would return error, because browser cannot listen.
func (nng *WASMMangosTransport) NewListener(addr string, sock mangos.Socket) (mangos.PipeListener, error) {
	return nil, errors.New("Cannot support listen in WASM")
}

// Scheme would be ws for WebSocket.
func (nng *WASMMangosTransport) Scheme() string {
	return "ws"
}

// WASMDialerWST dials websocket.
type WASMDialerWST struct {
	sock mangos.Socket
	url  string
}

// NewDialer creates WASMDialerWST.
func (nng *WASMMangosTransport) NewDialer(url string, sock mangos.Socket) (mangos.PipeDialer, error) {
	return &WASMDialerWST{
		sock: sock,
		url:  url,
	}, nil
}

// WASMPipeWS works with websocket connection.
type WASMPipeWS struct {
	open  bool
	conn  *net.Conn
	proto mangos.Protocol
}

// Send sends a complete message to websocket.
func (pipe *WASMPipeWS) Send(msg *mangos.Message) error {
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
func (pipe *WASMPipeWS) Recv() (*mangos.Message, error) {
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
func (pipe *WASMPipeWS) Close() error {
	pipe.open = false
	return (*pipe.conn).Close()
}

// LocalProtocol returns the 16-bit SP protocol number used by the
// local side.
func (pipe *WASMPipeWS) LocalProtocol() uint16 {
	return pipe.proto.Number()
}

// RemoteProtocol returns the 16-bit SP protocol number used by the
// remote side.  This will normally be received from the peer during
// connection establishment.
func (pipe *WASMPipeWS) RemoteProtocol() uint16 {
	return pipe.proto.PeerNumber()
}

// IsOpen returns true if the underlying connection is open.
func (pipe *WASMPipeWS) IsOpen() bool {
	return pipe.open
}

// GetProp returns an arbitrary transport specific property.
// These are like options, but are read-only and specific to a single
// connection. If the property doesn't exist, then ErrBadProperty
// should be returned.
func (pipe *WASMPipeWS) GetProp(string) (interface{}, error) {
	return nil, mangos.ErrBadProperty
}

// Dial is used to initiate a connection to a remote peer.
// It would open websocket connection to url specified in
// dialer.
func (dialer *WASMDialerWST) Dial() (mangos.Pipe, error) {
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
	return &WASMPipeWS{
		proto: dialer.sock.GetProtocol(),
		conn:  &conn,
		open:  true,
	}, nil
}

// SetOption sets a local option on the dialer.
// ErrBadOption can be returned for unrecognized options.
// ErrBadValue can be returned for incorrect value types.
func (dialer *WASMDialerWST) SetOption(name string, value interface{}) error {
	return mangos.ErrBadOption
}

// GetOption gets a local option from the dialer.
// ErrBadOption can be returned for unrecognized options.
func (dialer *WASMDialerWST) GetOption(name string) (value interface{}, err error) {
	return nil, mangos.ErrBadOption
}
