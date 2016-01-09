package BattleEye

import (
	"testing"
	"time"
)

func Test_StartAndStop(t *testing.T) {
	be := battleEye{}
	be.host = "127.000.000.001"
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
