package BattleEye

var packetType = struct {
	LOGIN          byte
	COMMAND        byte
	SERVER_MESSAGE byte
}{
	LOGIN:          0x00,
	COMMAND:        0x01,
	SERVER_MESSAGE: 0x02,
}
var packetResponse = struct {
	LOGIN_PASS           byte
	LOGIN_FAIL           byte
	MULTI_COMMAND_PACKET byte
}{
	LOGIN_PASS:           0x01,
	LOGIN_FAIL:           0x00,
	MULTI_COMMAND_PACKET: 0x00,
}
