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

	var tests = []struct {
		test     []byte
		expected byte
	}{
		{
			test:     []byte{'B', 'E', 1, 1, 1, 1, 0, 1, 255, 10, 5, 2, 82},
			expected: 255,
		},
		{
			test:     []byte{'B', 'E', 1, 1, 1, 1, 0, 1, 85},
			expected: 85,
		},
	}

	for _, v := range tests {
		result, err := getSequenceFromPacket(v.test)
		if err != nil {
			t.Error("Test:", v.test, "Failed due to error:", err)
		}
		if result != v.expected {
			t.Error("Expected:", v.expected, "Got:", result)
		}
	}

}
