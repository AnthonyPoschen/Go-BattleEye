package BattleEye

import (
	"hash/crc32"
	"testing"
)

func Test_getCheckSum(t *testing.T) {

}

func Test_dataMatchesCheckSum(t *testing.T) {
	var TestData = []struct {
		Data     []byte
		CheckSum uint32
	}{
		{[]byte("sjvbasdskbsdvj124582"), crc32.ChecksumIEEE([]byte("sjvbasdskbsdvj124582"))},
		{[]byte("0Password"), crc32.ChecksumIEEE([]byte("0Password"))},
	}

	for _, v := range TestData {
		if dataMatchesCheckSum(v.Data, v.CheckSum) != true {
			t.Error("test Data Failed")
		}
	}
}
