package BattleEye

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"testing"
)

func Test_getCheckSumFromBEPacket(t *testing.T) {

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
