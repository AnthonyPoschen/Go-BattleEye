// Package BattleEye doco goes here
package BattleEye

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

// Config File
type BattleEyeConfig struct {
	addr     *net.UDPAddr
	Password string
	// time in seconds to wait for a response. defaults to 2.
	ConnTimeout uint32
	// time in seconds to wait for response. defaults to 1.
	ResponseTimeout uint32
	// Time in seconds between sending a heartbeat when no commands are being sent. defaults 5
	HeartBeatTimer uint32
}

func (bec BattleEyeConfig) GetConfig() BattleEyeConfig {
	return bec
}

type BeConfig interface {
	GetConfig() BattleEyeConfig
}

//--------------------------------------------------

// This Struct holds the State of the connection and all commands
type battleEye struct {

	// Passed in config

	password        string
	addr            *net.UDPAddr
	connTimeout     uint32
	responseTimeout uint32
	heartbeatTimer  uint32
	// Sequence byte to determine the packet we are up to in the chain.
	sequence   byte
	chatWriter *io.Writer

	conn              *net.UDPConn
	lastCommandPacket time.Time
	running           bool
}

// Creates and Returns a new Client
func New(config BeConfig) *battleEye {
	// setup all variables
	cfg := config.GetConfig()
	if cfg.ConnTimeout == 0 {
		cfg.ConnTimeout = 2
	}
	if cfg.ResponseTimeout == 0 {
		cfg.ResponseTimeout = 1
	}
	if cfg.HeartBeatTimer == 0 {
		cfg.HeartBeatTimer = 5
	}

	return &battleEye{
		password:        cfg.Password,
		addr:            cfg.addr,
		connTimeout:     cfg.ConnTimeout,
		responseTimeout: cfg.ResponseTimeout,
		heartbeatTimer:  cfg.HeartBeatTimer,
	}
}

// Not Implemented
func (be *battleEye) SendCommand(command []byte) error {
	return nil
}

func (be *battleEye) Connect() (bool, error) {
	var err error
	// dial the Address
	be.conn, err = net.DialUDP("udp", nil, be.addr)
	if err != nil {
		return false, err
	}
	// make a buffer to read the packet packed with extra space
	packet := make([]byte, 9)

	// set timeout deadline so we dont block forever
	be.conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	// Send a Connection Packet
	be.conn.Write(buildConnectionPacket(be.password))
	// Read connection and hope it doesn't time out and the server responds
	n, err := be.conn.Read(packet)
	if err != nil {
		return false, err
	}

	result, err := checkLogin(packet[:n])
	if err != nil {
		return false, err
	}
	if result == packetResponse.LOGIN_FAIL {
		return false, nil
	}
	// nothing has failed we are good to go :).
	return true, nil
}

func (be *battleEye) Disconnect() error {
	// maybe also close the main loop and wait for that?
	be.conn.Close()
	return nil
}

func checkLogin(packet []byte) (byte, error) {
	var err error = nil
	if len(packet) != 9 {
		return 0, errors.New("Packet Size Invalid for Response")
	}
	// check if we have a valid packet
	if match, err := PacketMatchesChecksum(packet); match == false || err != nil {
		return 0, err
	}
	// now check if we got a success or a fail
	// 2 byte prefix. 4 byte checksum. 1 byte terminate header. 1 byte login type. 1 byte result
	return packet[8], err
}

func (be *battleEye) ReadPacketLoop(ch chan []byte) {
	data := make([]byte, 4096)
	be.conn.SetReadBuffer(4096)
	for {
		be.conn.SetReadDeadline(time.Now().Add(time.Second))
		count, err := be.conn.Read(data)
		fmt.Println("Read:", data[:count])
		if err != nil {
			fmt.Println("Failed to read from connection:", err)
			return
		}
		fmt.Println("Recieved Packet")
		go func() { ch <- data[:count] }()
		return
	}
}
