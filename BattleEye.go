// Package BattleEye doco goes here
package BattleEye

import (
	"errors"
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
	PollFreq struct {
		Chat int
		Bans int
	}
}

func (bec BattleEyeConfig) GetConfig() BattleEyeConfig {
	return bec
}

type BeConfig interface {
	GetConfig() BattleEyeConfig
}

type ban struct {
	GUID     string
	Duration uint32
	Reason   string
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

	finish   chan struct{}
	packet   chan []byte
	done     sync.WaitGroup
	conn     *net.UDPConn
	running  bool
	sequence byte
	chat     []string
	bans     []ban
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
	BE.chat = make([]string, 0)
	return &BE
}

func (be *battleEye) SendCommand(command []byte) error {
	return nil
}

// returns a copy of the current chat that has been sent from battlEye since the last clear.
func (be *battleEye) GetChat() ([]string, error) {
	if be.chat == nil {
		return []string{}, errors.New("Chat not initilised")
	}
	return be.chat, nil
}

// Clears the chat. this becomes very useful for any server wanting to perform information based on chat responses.
// once procesed you can then remove the messages by calling clear.
func (be *battleEye) ClearChat() error {
	if be.chat == nil {
		return errors.New("Chat not initilised")
	}
	be.chat = nil
	return nil
}

func (be *battleEye) GetBans() ([]ban, error) {
	if be.bans == nil {
		return []ban{}, errors.New("Bans not initilised")
	}
	return be.bans, nil
}
