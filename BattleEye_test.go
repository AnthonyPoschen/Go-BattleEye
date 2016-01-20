package BattleEye

import (
	"encoding/binary"
	"testing"
)

func Test_StartAndStop(t *testing.T) {
	/*
		be := battleEye{}
		be.host = "127.0.0.1"
		be.port = "1000"
		go be.Run()
		finished := make(chan struct{})
		go func() {
			be.Stop()
			finished <- struct{}{}
		}()
		select {
		case <-finished:

		case <-time.After(time.Millisecond * 500):
			t.Error("failed to Stop in appropiate time")
		}
	*/
}

func Test_BuildHeader(t *testing.T) {
	TestValues := []uint32{
		58,
		25,
		1400,
		980,
		4294967295,
		0,
		2147483647,
		1600581284,
		3848910246,
		108500257,
	}
	a := make([]byte, 4)
	for _, v := range TestValues {
		binary.LittleEndian.PutUint32(a, v)
		h := buildHeader(v)
		if len(h) != 6 {
			t.Error("Header Invalid Size")
		}
		if h[0] != 'B' || h[1] != 'E' || h[2] != a[0] || h[3] != a[1] || h[4] != a[2] || h[5] != a[3] {
			t.Error("Header Signature Not Correct")
		}
	}
}

func Test_processPacket(t *testing.T) {

	var tests = []struct {
		Packet   []byte
		Sequence byte
	}{
		{
			Packet:   buildCommandPacket([]byte{'f', 'u'}, 0x32),
			Sequence: 0x32,
		},
	}

	be := BattleEye{}
	for _, v := range tests {
		be.packetQueue.queue = append(be.packetQueue.queue, transmission{packet: v.Packet, sequence: v.Sequence})
	}
	if len(be.packetQueue.queue) != len(tests) {
		t.Error("BattlEye queue length wrong expected:", len(tests), "Got:", len(be.packetQueue.queue))
	}

	for _, v := range tests {
		err := be.processPacket(v.Packet)
		if err != nil {
			t.Error(err)
		}
	}

	if len(be.packetQueue.queue) != 0 {
		t.Error("Packet Queue not 0", be.packetQueue.queue)
	}
}
