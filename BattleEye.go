// Package BattleEye doco goes here
package BattleEye

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// ErrTimeout is a timeout error and should be seen as
// failing to establish a connection
var ErrTimeout = errors.New("Connection Timed Out")

var IoDiscard struct {
	io.WriteCloser
}

// Config documentation
type Config struct {
	Addr     *net.UDPAddr
	Password string
	// time in seconds to wait for a response. defaults to 2.
	ConnTimeout uint32
	// time in seconds to wait for response. defaults to 1.
	ResponseTimeout uint32
	// wait time after first command response to check if multiple packets arrive.
	// defaults. 0.5s
	MultiResponseTimeout uint32
	// Time in seconds between sending a heartbeat when no commands are being sent. defaults 5
	HeartBeatTimer uint32
}

// GetConfig Returns the config to satisfy the interface for setting up a new battleeye connection
func (bec Config) GetConfig() Config {
	return bec
}

// BeConfig is the interface for passing in a configuration for the client.
// this allows other types to be implemented that also contain the type desired
type BeConfig interface {
	GetConfig() Config
}
type transmission struct {
	packet   []byte
	sequence byte
	response []byte
	sent     time.Time
	counter  int
	w        io.WriteCloser
}

//--------------------------------------------------

// BattleEye must do doco soon
type BattleEye struct {

	// Passed in config

	password             string
	addr                 *net.UDPAddr
	connTimeout          uint32
	responseTimeout      uint32
	multiResponseTimeout uint32
	heartbeatTimer       uint32
	reconnectTimeOut     float64
	timeofLastPacket     time.Time
	connected            bool

	// Sequence byte to determine the packet we are up to in the chain.
	sequence struct {
		sync.Mutex
		n byte
	}
	chatWriter struct {
		sync.Mutex
		io.Writer
	}
	writebuffer []byte

	conn              *net.UDPConn
	lastCommandPacket struct {
		sync.Mutex
		time.Time
	}
	running bool
	wg      sync.WaitGroup

	// use this to unlock before reading.
	// and match reads to waiting confirms to purge this list.
	// or possibly resend
	packetQueue struct {
		sync.Mutex
		queue []transmission
	}
}

// New Creates and Returns a new Client
func New(config BeConfig) *BattleEye {
	// setup all variables
	cfg := config.GetConfig()
	if cfg.ConnTimeout == 0 {
		cfg.ConnTimeout = 2
	}
	if cfg.ResponseTimeout == 0 {
		cfg.ResponseTimeout = 2
	}
	if cfg.HeartBeatTimer == 0 {
		cfg.HeartBeatTimer = 5
	}

	return &BattleEye{
		password:         cfg.Password,
		addr:             cfg.Addr,
		connTimeout:      cfg.ConnTimeout,
		responseTimeout:  cfg.ResponseTimeout,
		heartbeatTimer:   cfg.HeartBeatTimer,
		writebuffer:      make([]byte, 4096),
		reconnectTimeOut: 40,
		wg:               sync.WaitGroup{},
	}

}

func (be *BattleEye) getSequence() byte {
	be.sequence.Lock()
	defer be.sequence.Unlock()
	sequence := be.sequence.n
	// increment the sending packet.
	if be.sequence.n == 255 {
		be.sequence.n = 0
	} else {
		be.sequence.n++
	}
	return sequence
}

// SendCommand takes a byte array of a command string i.e 'ban xyz' and a io.Writer, it will
// Make sure the server recieves the command by retrying if needed and write the response to the writer.
// if no response is recieved then a response has not yet been recieved. a empty write
func (be *BattleEye) SendCommand(command []byte, w io.WriteCloser) error {
	sequence := be.getSequence()
	packet := buildCommandPacket(command, sequence)
	fmt.Println("Sending packet Seq:", sequence, "command:", string(command))
	return be.sendPacket(packet, w, 0)
}

func (be *BattleEye) sendPacket(packet []byte, w io.WriteCloser, counter int) error {
	sequence, err := getSequenceFromPacket(packet)

	if err != nil {

		return err
	}
	if be.connected {
		fmt.Println("Packet:", packet, "Sequence:", sequence, "err:", err)
		be.conn.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(be.responseTimeout)))
		_, err = be.conn.Write(packet)
		be.packetQueue.Lock()
		be.packetQueue.queue = append(be.packetQueue.queue, transmission{packet: packet, sequence: sequence, sent: time.Now(), w: w, counter: counter})
		be.packetQueue.Unlock()
		be.lastCommandPacket.Lock()
		be.lastCommandPacket.Time = time.Now()
		be.lastCommandPacket.Unlock()
	}

	return err
}

// Connect attempts to establish a connection with the BattlEye Rcon server and if it works it then sets up a loop in a goroutine
// to recieve all callbacks.
func (be *BattleEye) Connect() (bool, error) {
	be.wg = sync.WaitGroup{}
	var err error
	// dial the Address

	be.conn, err = net.DialUDP("udp", nil, be.addr)
	if err != nil {
		fmt.Println("Fail dial")
		return false, err
	}
	// make a buffer to read the packet packed with extra space
	buf := make([]byte, 20)
	// set timeout deadline so we dont block forever
	be.conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	// Send a Connection Packet
	be.conn.SetWriteDeadline(time.Now().Add(time.Second * 2))
	_, err = be.conn.Write(buildConnectionPacket(be.password))
	if err != nil {
		return false, err
	}
	// Read connection and hope it doesn't time out and the server responds
	n, err := be.conn.Read(buf)

	// check if this is a timeout error.
	if err, ok := err.(net.Error); ok && err.Timeout() {
		//fmt.Println("error", err)
		return false, ErrTimeout
	}
	if err != nil {
		return false, err
	}
	fmt.Println(buf[:n])
	result, err := checkLogin(buf[:n])
	if err != nil {
		return false, err
	}

	if result == packetResponse.LoginFail {
		return false, nil
	}

	// nothing has failed we are good to go :).
	// Spin up a go routine to read back on a connection

	be.wg.Add(1)
	be.timeofLastPacket = time.Now()
	be.lastCommandPacket.Lock()
	be.lastCommandPacket.Time = time.Now()
	be.lastCommandPacket.Unlock()
	go be.updateLoop()
	fmt.Println("Succesfully Logged into Rcon")
	be.connected = true
	return true, nil
}

func (be *BattleEye) updateLoop() {
	defer be.wg.Done()
	for {
		if be.conn == nil {
			fmt.Println("Connection is NIL")
			return
		}
		t := time.Now()
		if t.Sub(be.timeofLastPacket).Seconds() > be.reconnectTimeOut {
			fmt.Println("Reconnecting")
			be.connected = false
			b, err := be.Reconnect()
			if b {
				be.timeofLastPacket = time.Now()
				be.connected = true
				fmt.Println("Reconnected")
			} else {
				fmt.Println("Failed to reconnect reason:", err)
				<-time.After(time.Second * 15)
			}
		}

		be.lastCommandPacket.Lock()
		if t.After(be.lastCommandPacket.Add(time.Second * time.Duration(be.heartbeatTimer))) {
			go be.SendCommand([]byte{}, IoDiscard)
			be.lastCommandPacket.Time = t
		}
		be.lastCommandPacket.Unlock()

		go be.checkOldPackets()
		// do check for new incoming data
		be.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 20))
		n, err := be.conn.Read(be.writebuffer)
		if err != nil {
			continue
		}
		go be.processPacket(be.writebuffer[:n])

	}
}
func (be *BattleEye) processPacket(data []byte) {
	be.timeofLastPacket = time.Now()
	sequence, content, pType, err := verifyPacket(data)
	if err != nil {
		fmt.Println("Read Packet Error:", err)
		return
	}
	fmt.Println("Recieved packet with Seq", int(sequence), "Content:", string(content))
	// if say command write acknoledge and leave
	if pType == packetType.ServerMessage {
		be.sequence.Lock()
		if sequence == 255 {
			be.sequence.n = 0x00
		} else {
			be.sequence.n = sequence + 0x01
		}
		//fmt.Println("incremented Seq to ", be.sequence.n)
		be.sequence.Unlock()

		be.checkSequenceClash(sequence)
		be.chatWriter.Lock()
		if be.chatWriter.Writer != nil {
			be.chatWriter.Write(append(content[3:], []byte{'\n'}...))
		}
		be.chatWriter.Unlock()
		// we must acknoledge we recieved this first
		if be.conn != nil {
			be.conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 100))
			_, err := be.conn.Write(buildPacket([]byte{sequence}, packetType.ServerMessage))
			if err != nil {
				fmt.Println(err)
			}
		}

		return
	}

	// else for command check if we expect more packets and how many.
	if pType != packetType.Command && pType != 0x00 {
		fmt.Println("Unknown way to respond to packet type: " + string(pType))
		return
	}
	packetCount, currentPacket, isMultiPacket := checkMultiPacketResponse(content[1:])
	//fmt.Println("sequence", sequence)
	// process the packet if it is not a multipacket
	if !isMultiPacket {
		//fmt.Println("returning", sequence)
		be.handleResponseToQueue(sequence, content[3:], false)
		return
	}

	if currentPacket+1 < packetCount {
		//fmt.Println("add")
		be.handleResponseToQueue(sequence, content[6:], true)
	} else {
		//fmt.Println("finish")
		be.handleResponseToQueue(sequence, content[6:], false)
	}
	return
	// loop till we have all the messages and i guess send confirms back.
	/*for ; currentPacket >= packetCount; currentPacket++ {
		//fmt.Println("Is Multi Packet. Expecting", packetCount, "Current Packet", currentPacket)
		fmt.Println("currentPacket", currentPacket)
		be.conn.SetReadDeadline(time.Now().Add(time.Second * 3))
		n, err := be.conn.Read(be.writebuffer)
		if err != nil {
			fmt.Println(err)
			break
		}
		// lets re verify this entire thing
		p := be.writebuffer[:n]
		seq, cont, _, err := verifyPacket(p)
		if err != nil {
			return err
		}
		if seq != sequence {
			fmt.Println("rehandling packet")
			be.processPacket(p)
			currentPacket--
		}
		be.handleResponseToQueue(seq, cont[5:], true)

	}
	//closes it as we know we have done our job
	fmt.Println("finished", sequence)
	be.handleResponseToQueue(sequence, []byte{}, false)
	// now that we have goten all the packets we are after and writen them to the buffer lets return the result.
	return nil
	*/
}

// Disconnect shuts down the infinite loop to recieve packets and closes the connection
func (be *BattleEye) Disconnect() error {
	// maybe also close the main loop and wait for that?
	be.conn.Close()
	be.wg.Wait()
	return nil
}

func (be *BattleEye) Reconnect() (bool, error) {
	packet := make([]byte, 9)
	be.conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	// Send a Connection Packet
	be.conn.SetWriteDeadline(time.Now().Add(time.Second * 2))
	be.conn.Write(buildConnectionPacket(be.password))
	// Read connection and hope it doesn't time out and the server responds
	n, err := be.conn.Read(packet)
	// check if this is a timeout error.
	if err, ok := err.(net.Error); ok && err.Timeout() {
		return false, ErrTimeout
	}
	if err != nil {
		return false, err
	}

	result, err := checkLogin(packet[:n])
	if err != nil {
		return false, err
	}

	if result == packetResponse.LoginFail {
		return false, nil
	}
	be.timeofLastPacket = time.Now()
	be.connected = true
	return true, nil
}

//SetChatWriter Sets a writer to write all chat messages too as they come in
func (be *BattleEye) SetChatWriter(w io.Writer) {
	be.chatWriter.Lock()
	be.chatWriter.Writer = w
	be.chatWriter.Unlock()
}
func (be *BattleEye) checkSequenceClash(sequence byte) {
	be.packetQueue.Lock()
	defer be.packetQueue.Unlock()
	//fmt.Println("Packet Queue Size", len(be.packetQueue.queue))
	for i := 0; i < len(be.packetQueue.queue); i++ {
		v := be.packetQueue.queue[i]
		if v.sequence == sequence {
			be.packetQueue.queue = append(be.packetQueue.queue[:i], be.packetQueue.queue[i+1:]...)
			i--
			//packet, _ := replaceSequence(v.packet, be.getSequence())
			//fmt.Println("Packet Clash Resending:", v.packet)
			//go be.sendPacket(packet, v.w, v.counter)
		}

	}
}
func checkLogin(packet []byte) (byte, error) {
	var err error
	if len(packet) != 9 {
		return 0, errors.New("Packet Size Invalid for Response")
	}
	// check if we have a valid packet
	if match, err := packetMatchesChecksum(packet); match == false || err != nil {
		return 0, err
	}
	// now check if we got a success or a fail
	// 2 byte prefix. 4 byte checksum. 1 byte terminate header. 1 byte login type. 1 byte result
	return packet[8], err
}

func (be *BattleEye) handleResponseToQueue(sequence byte, response []byte, moreToCome bool) {
	be.packetQueue.Lock()
	for k, v := range be.packetQueue.queue {
		if v.sequence == sequence {
			if v.w != nil {
				extra := []byte("\n")
				if moreToCome {
					extra = []byte{}
				}

				//v.response = append(v.response,response...)
				//v.response = append(v.response,extra...)

				//if !moreToCome {
				v.w.Write(append(response, extra...))
				//}

			}

			if !moreToCome {
				if v.w != nil {
					v.w.Close()
				}

				be.packetQueue.queue = append(be.packetQueue.queue[:k], be.packetQueue.queue[k+1:]...)

			}
			break
		}
	}
	be.packetQueue.Unlock()
}

func verifyPacket(data []byte) (sequence byte, content []byte, pType byte, err error) {
	checksum, err := getCheckSumFromBEPacket(data)
	if err != nil {
		return
	}
	if len(data) < 6 {
		err = errors.New("Packet size too small to have data")
		return
	}
	match := dataMatchesCheckSum(data[6:], checksum)
	if !match {
		err = errors.New("Checksum does not match data")
		return
	}
	sequence, err = getSequenceFromPacket(data)
	if err != nil {
		return
	}

	content, err = stripHeader(data)
	if err != nil {
		return
	}
	pType, err = responseType(data)

	return
}

func (be *BattleEye) checkOldPackets() {
	be.packetQueue.Lock()
	defer be.packetQueue.Unlock()
	t := time.Now()
	for i := 0; i < len(be.packetQueue.queue); i++ {
		v := be.packetQueue.queue[i]
		if t.After(v.sent.Add(time.Second * 2)) {
			//fmt.Println(string(v.packet), "Count:", v.counter)
			v.counter++
			be.packetQueue.queue = append(be.packetQueue.queue[:i], be.packetQueue.queue[i+1:]...)
			if v.counter < 4 && t.After(v.sent.Add(time.Second*2)) {
				//seq := be.getSequence()
				//packet, _ := replaceSequence(v.packet, seq)
				//fmt.Println("Sending new packet seq:", int(seq), "packet:", packet)
				//go be.sendPacket(packet, v.w, v.counter)
			}
			i--
		}
	}
}
