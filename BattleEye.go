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

	finish chan struct{}
	done   sync.WaitGroup
	conn   *net.UDPConn
}

// Blocking function will not return until closed with Stop.
func (be *battleEye) Run() {
	// setup for running
	be.finish = make(chan struct{})
	be.done.Add(1)
	defer be.done.Done()

	// need to setup actual work channels.
	var err error
	var addr *net.UDPAddr
	addr, err = net.ResolveUDPAddr("udp", be.host+":"+be.port)
	be.conn, err = net.DialUDP("udp4", nil, addr)

	if err != nil {
		fmt.Println("error Connecting:", err)
		return
	}
	defer be.conn.Close()
	// dont set Update packet yet so it stays blocking till we connect
	var UpdatePacket <-chan time.Time
	ConnectTimer := time.After(time.Millisecond)

loop:
	for {

		select {
		// need to create case's for doing actual work.
		case <-ConnectTimer:
			// send a connect packet
			//be.conn.Write()
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
	fmt.Println(be.conn)
}

// Creates and Returns a new Client
func New(config BeConfig) *battleEye {
	// setup all variables
	cfg := config.GetConfig()
	BE := battleEye{host: cfg.Host, port: cfg.Port, password: cfg.Password}
	return &BE
}

func processPacket(data []byte) {

}

//func acknoledgePacket

func buildHeader(Checksum uint32) []byte {
	Check := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(Check, Checksum)
	// build header and return it.
	return append([]byte{}, 'B', 'E', Check[0], Check[1], Check[2], Check[3], 0xFF)
}
