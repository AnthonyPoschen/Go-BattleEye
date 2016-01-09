// Package BattleEye doco goes here
package BattleEye

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

// Config File
type BattleEyeConfig struct {
	Host     string
	Port     string
	Password string
}

func (bec *BattleEyeConfig) GetConfig() struct {
	Host     string
	Port     string
	Password string
} {
	return struct {
		Host     string
		Port     string
		Password string
	}{
		bec.Host,
		bec.Port,
		bec.Password,
	}
}

type BeConfig interface {
	GetConfig() struct {
		Host     string
		Port     string
		Password string
	}
}

//--------------------------------------------------

// This Struct holds the State of the connection and all commands
type battleEye struct {
	// IP Address to connect too
	host string
	// Port to Connect too
	port string
	// RCON Password
	password string

	finish  chan struct{}
	packet  chan []byte
	done    sync.WaitGroup
	conn    *net.UDPConn
	running bool
}

// Blocking function will not return until closed with Stop.
func (be *battleEye) Run() {
	// setup for running
	be.finish = make(chan struct{})
	be.packet = make(chan []byte)
	defer close(be.packet)
	be.done.Add(1)
	defer be.done.Done()

	// need to setup actual work channels.
	var err error
	var addr *net.UDPAddr

	addr, err = net.ResolveUDPAddr("udp", be.host+":"+be.port)

	be.conn, err = net.DialUDP("udp", nil, addr)

	if err != nil {
		fmt.Println("error Connecting:", err)
		return
	}
	defer be.conn.Close()
	// dont set Update packet yet so it stays blocking till we connect
	var UpdatePacket <-chan time.Time
	ConnectTimer := time.After(time.Millisecond)
	be.running = true
	newPacketArrived := make(chan []byte)
	be.conn.SetWriteBuffer(4096)
	//be.conn.WriteMsgUDP(b, oob, addr)
loop:
	for {

		select {
		case packet := <-newPacketArrived:
			fmt.Println(packet)
		// need to create case's for doing actual work.
		case <-ConnectTimer:
			// send a connect packet
			be.conn.Write(buildConnectionPacket(be.password))
			be.ReadPacketLoop(newPacketArrived)
		case <-UpdatePacket:
			UpdatePacket = time.After(time.Second * 5)
			// do Update Packet
		// if called to exit do the following.
		case _, ok := <-be.finish:
			if ok == false {
				// signal services to break
				// i never knew you could break labels - Quality of life improved
				break loop
			}
		}

	}

	// signals we are finished

}

// Will stop the running connection
func (be *battleEye) Stop() {
	close(be.finish)
	be.done.Wait()
	be.running = false
}

func (be *battleEye) IsRunning() bool {
	return be.running
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

// Creates and Returns a new Client
func New(config BeConfig) *battleEye {
	// setup all variables
	cfg := config.GetConfig()
	BE := battleEye{host: cfg.Host, port: cfg.Port, password: cfg.Password}
	return &BE
}

func buildHeader(Checksum uint32) []byte {
	Check := make([]byte, 4) // should reduce allocations when i benchmark this shit
	binary.LittleEndian.PutUint32(Check, Checksum)
	// build header and return it.
	return append([]byte{}, 'B', 'E', Check[0], Check[1], Check[2], Check[3])
}

func buildConnectionPacket(pass string) []byte {
	data := append([]byte{255, 0}, []byte(pass)...)
	checksum := makeChecksum(data)
	binary.LittleEndian.Uint32([]byte{37, 111, 118, 65})
	header := buildHeader(checksum)
	return append(header, data...)
}
