package BattleEye

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"testing"
)

func Test_getCheckSumFromBEPacket(t *testing.T) {
	var tests = []struct {
		test     []byte
		expected uint32
	}{
		{
			test:     []byte{'B', 'E', 0, 1, 2, 3, 0xFF},
			expected: (uint32(0) | uint32(1)<<8 | uint32(2)<<16 | uint32(3)<<24),
		},
	}

	for _, v := range tests {
		res, err := getCheckSumFromBEPacket(v.test)
		if err != nil {
			t.Error(err)
		}
		if res != v.expected {
			t.Error("Expected:", v.expected, "Recieved:", res)
		}
	}
}

func Test_dataMatchesCheckSum(t *testing.T) {
	var TestData = []struct {
		Data     []byte
		CheckSum uint32
	}{
		{[]byte("sjvbasdskbsdvj124582"), crc32.ChecksumIEEE([]byte("sjvbasdskbsdvj124582"))},
		{[]byte("admin"), crc32.ChecksumIEEE([]byte("admin"))},
	}
	for _, v := range TestData {
		if dataMatchesCheckSum(v.Data, v.CheckSum) != true {
			t.Error("test Data Failed")
		}

	}
}

func Test_RealByteCRCCheck(t *testing.T) {

	// Confirmed IEEE CRC
	// and small endian
	//this is an actual Result from a live server
	hash := []byte{37, 111, 118, 65}
	data := []byte{255, 0, 97, 100, 109, 105, 110}
	RealHash := binary.LittleEndian.Uint32(hash)
	ActualHash := crc32.Checksum(data, crc32.MakeTable(crc32.IEEE))
	liveExample := makeChecksum(data)
	livebinary := buildConnectionPacket("admin")[2:7]
	LivePacket := binary.LittleEndian.Uint32(livebinary)

	if RealHash != ActualHash {
		t.Error("Example Hash Does Not Match")
	}
	if RealHash != liveExample {
		t.Error("Hash Is not correctly Calculated in makeChecksum")
	}
	if RealHash != LivePacket {
		t.Error("Hash is not correctly Stored in Connection Packet\nExpected:", hash, "\nRecieved:", livebinary)
	}
	fmt.Println()
}
