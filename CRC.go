package BattleEye

import (
	"errors"
	"hash/crc32"
)

var crcTable *crc32.Table = crc32.MakeTable(crc32.IEEE)

func getCheckSumFromBEPacket(data []byte) (uint32, error) {
	notValidString := "Data not a Valid BE HEader: "
	// check the data is minimum the size of a BE header
	if len(data) != 7 {
		return 0, errors.New(notValidString + "Header Size Not Valid")
	}
	if data[0] != 'B' || data[1] != 'E' {
		return 0, errors.New(notValidString + "'BE' Not at Start of Header")
	}
	if data[6] != 0xff {
		return 0, errors.New(notValidString + "Header does not end with '0xff'")
	}

	// little endian uint32. lets hope its fucking correct fucking battleeye not listing shit in protocol
	// it should because other tools are in little i had to go verify this ffs.
	result := uint32(data[3]) | uint32(data[4])<<8 | uint32(data[5])<<16 | uint32(data[6])<<24
	return result, nil
}

func dataMatchesCheckSum(data []byte, Checksum uint32) bool {
	return crc32.Checksum(data, crcTable) == Checksum

}

func makeChecksum(data []byte) uint32 {
	return crc32.Checksum(data, crcTable)
}
