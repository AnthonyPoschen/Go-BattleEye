package BattleEye

import "testing"

func Test_stripHeader(t *testing.T) {
	cmd := "Ban randomUser 0"
	packet := buildPacket([]byte(cmd), packetType.Command)
	result, err := stripHeader(packet)
	if err != nil {

		t.Fatal("Error on StripHeader:", err.Error())
	}
	if string(result) != cmd {
		t.Fatal("Expected:", cmd, "Got:", string(result))
	}
}

func Test_getSequenceFromPacket(t *testing.T) {
	sequence := byte(255)

	tests := []struct{test []byte,
		expected byte,
	} {
		{

			},
			{

			}
	}
	packet := []byte{'B','E',1,1,1,1,0,1,sequence,10,5,2,82}
	packet2 := []byte{'B','E',1,1,1,1,0,1,sequence}

	result , err := getSequenceFromPacket(packet)
	if err != nil
}
