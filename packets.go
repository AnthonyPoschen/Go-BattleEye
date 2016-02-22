package BattleEye

import (
	"encoding/binary"
	"errors"
)

func buildHeader(Checksum uint32) []byte {
	Check := make([]byte, 4) // should reduce allocations when i benchmark this shit
	binary.LittleEndian.PutUint32(Check, Checksum)
	// build header and return it.
	return append([]byte{}, 'B', 'E', Check[0], Check[1], Check[2], Check[3])
}

//This Takes a Constructed command or login packet and wraps it with the header and assigned a checksum
func buildPacket(data []byte, PacketType byte) []byte {
	data = append([]byte{0xFF, PacketType}, data...)
	checksum := makeChecksum(data)
	header := buildHeader(checksum)
	return append(header, data...)
}

func buildConnectionPacket(pass string) []byte {
	return buildPacket([]byte(pass), packetType.Login)
}

func buildCommandPacket(command []byte, sequence uint8) []byte {
	return buildPacket(append([]byte{sequence}, command...), packetType.Command)
}

// Note sure if this command packet heartbeat needs to keep track of sequence or not. but thanks battle eye ill presume
// it does since it asks for a 2 byte empty command
func buildHeartBeatPacket(sequence uint8) []byte {
	return buildPacket([]byte{sequence}, packetType.Command)
}

func buildMessageAckPacket(sequence uint8) []byte {
	return buildPacket([]byte{sequence}, packetType.ServerMessage)
}

//This function takes in a data packet and returns the type of packet recieved
func responseType(data []byte) (byte, error) {
	if len(data) < 8 {
		return 0, errors.New("Error Packet length too small")
	}
	// 7th element will be the first element after the header which will be the packet type
	return data[7], nil
}

func stripHeader(data []byte) ([]byte, error) {
	if len(data) < 7 {
		return []byte{}, errors.New("Size to small, Non valid header")
	}

	return data[6:], nil
}

func stripHeaderAndCommand(data []byte) ([]byte, error) {
	if len(data) < 9 {
		return []byte{}, errors.New("Size to small, Non valid Header")
	}

	return data[8:], nil
}
func getSequenceFromPacket(data []byte) (byte, error) {
	if len(data) < 9 {
		return 0, errors.New("Packet to small no Sequence Stored")
	}
	return data[8], nil
}

// Returns TotalPackets , CurrentPacket , IsMultiPacket
func checkMultiPacketResponse(data []byte) (byte, byte, bool) {
	if len(data) < 3 {
		return 0, 0, false
	}
	if data[0] != 0x01 || data[2] != 0x00 {
		return 0, 0, false
	}
	return data[3], data[4], true
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

func replaceSequence(packet []byte, sequence byte) ([]byte, error) {
	if len(packet) < 10 {
		return []byte{}, errors.New("Packet to small")
	}
	packet[8] = sequence
	data, _ := stripHeader(packet)
	checksum := makeChecksum(data)
	header := buildHeader(checksum)
	return append(header, data...), nil
}
