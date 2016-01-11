package BattleEye

// Types of packets used to communicate to BattlEye
var packetType = struct {
	LOGIN          byte
	COMMAND        byte
	SERVER_MESSAGE byte
}{
	LOGIN:          0x00,
	COMMAND:        0x01,
	SERVER_MESSAGE: 0x02,
}

// Might rename this. but this is usually the identifier after a packet type
// to give more context. i.e when recieving a login packet from the server it
// will then send a pass or fail byte right after the packet type
var packetResponse = struct {
	LOGIN_PASS           byte
	LOGIN_FAIL           byte
	MULTI_COMMAND_PACKET byte
}{
	LOGIN_PASS:           0x01,
	LOGIN_FAIL:           0x00,
	MULTI_COMMAND_PACKET: 0x00,
}
