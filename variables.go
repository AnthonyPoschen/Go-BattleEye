package BattleEye

// Types of packets used to communicate to BattlEye
var packetType = struct {
	Login         byte
	Command       byte
	ServerMessage byte
}{
	Login:         0x00,
	Command:       0x01,
	ServerMessage: 0x02,
}

// Might rename this. but this is usually the identifier after a packet type
// to give more context. i.e when recieving a login packet from the server it
// will then send a pass or fail byte right after the packet type
var packetResponse = struct {
	LoginPass          byte
	LoginFail          byte
	MultiCommandPacket byte
}{
	LoginPass:          0x01,
	LoginFail:          0x00,
	MultiCommandPacket: 0x00,
}
