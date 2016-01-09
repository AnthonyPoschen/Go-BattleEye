package BattleEye

import (
	"encoding/binary"
	"testing"
	"time"
)

func Test_StartAndStop(t *testing.T) {
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
		if h[0] != 'B' || h[1] != 'E' || h[2] != a[0] || h[3] != a[1] || h[4] != a[2] || h[5] != a[3] || h[6] != 0xFF {
			t.Error("Header Signature Not Correct")
		}
	}
}

func Test_LiveServer(t *testing.T) {
	be := New(&BattleEyeConfig{Host: "127.0.0.1", Port: "2302", Password: "TestPw"})
	go be.Run()
	<-time.After(time.Second * 10)
	be.Stop()
}
