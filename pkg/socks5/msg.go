package socks5

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

const (
	AuthMethodNone = 0x00
)

type ClientMethodSelectionMessage struct {
	ver      uint8
	nmethods uint8
	methods  []uint8
}

func (m *ClientMethodSelectionMessage) FromReader(r io.Reader) error {
	b := make([]byte, 2)

	if _, err := io.ReadFull(r, b); err != nil {
		return err
	}

	m.ver, m.nmethods = b[0], b[1]
	if m.ver != Socks5Version {
		return errors.New("socks protocol version mismatch")
	}

	m.methods = make([]uint8, m.nmethods)

	_, err := io.ReadFull(r, m.methods)
	return err
}

type ServerMethodSelectionMessage struct {
	ver    uint8
	method uint8
}

// todo func (m *ServerMethodSelectionMessage) FromReader

func (m *ServerMethodSelectionMessage) Raw() []byte {
	return []byte{m.ver, m.method}
}

const (
	CmdConnect      = 0x01
	CmdBind         = 0x02
	CmdUdpAssociate = 0x03
)

type ClientRequestMessage struct {
	ver     uint8
	cmd     uint8
	rsv     uint8
	atyp    uint8
	dstAddr []byte
	dstPort uint16
}

func (m *ClientRequestMessage) FromReader(r io.Reader) error {
	b := make([]byte, 4)
	if _, err := io.ReadFull(r, b); err != nil {
		return err
	}
	m.ver, m.cmd, m.rsv, m.atyp = b[0], b[1], b[2], b[3]

	var err error
	m.dstAddr, err = readAddr(r, m.atyp)
	if err != nil {
		return err
	}

	b = make([]byte, 2)
	if _, err = io.ReadFull(r, b); err != nil {
		return err
	}

	m.dstPort = binary.BigEndian.Uint16(b)
	return nil
}

const (
	RepSuccess       = 0x00
	RepServerFailure = 0x01
	RepNotAllowed    = 0x02
)

type ServerReplyMessage struct {
	ver     uint8
	rep     uint8
	rsv     uint8
	atyp    uint8
	bndAddr []byte
	bndPort uint16
}

func (m *ServerReplyMessage) Raw() []byte {
	b := append([]byte{m.ver, m.rep, m.rsv, m.atyp}, m.bndAddr...)

	portBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(portBuf, m.bndPort)

	return append(b, portBuf...)
}

type UdpPacket struct {
	rsv     uint16
	frag    uint8
	atyp    uint8
	dstAddr []byte
	dstPort uint16
	data    []byte
}

func (m *UdpPacket) FromReader(r io.Reader) error {
	b := make([]byte, 4)
	if _, err := io.ReadFull(r, b); err != nil {
		return err
	}
	m.rsv = binary.BigEndian.Uint16(b[:2])

	m.frag, m.atyp = b[2], b[3]

	var err error
	m.dstAddr, err = readAddr(r, m.atyp)
	if err != nil {
		return err
	}

	b = make([]byte, 2)
	if _, err = io.ReadFull(r, b); err != nil {
		return err
	}

	m.dstPort = binary.BigEndian.Uint16(b)

	m.data, err = io.ReadAll(r)
	return err
}

func (m *UdpPacket) Raw() []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, m.rsv)

	b = append(b, m.frag, m.atyp)
	b = append(b, m.dstAddr...)

	portBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(portBuf, m.dstPort)

	b = append(b, portBuf...)
	return append(b, m.data...)
}
