package iafon

import (
	"net/http"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	if NewServer() == nil {
		t.Fatal("NewServer() == nil")
	}
}

func TestRunWithoutRoute(t *testing.T) {
	s := NewServer("127.0.0.1:")
	err := s.Run()

	if err == nil {
		t.Fatal("Run before add routes should return error")
	}
}

func TestRunWithRoute(t *testing.T) {
	s := NewServer("127.0.0.1:")
	s.Handle("GET", "/", func(http.ResponseWriter, *http.Request) {})

	go func() {
		time.Sleep(time.Millisecond)
		time.Sleep(time.Millisecond)
		s.Close()
	}()

	err := s.Run()

	if err != nil && err.Error() != "http: Server closed" {
		t.Fatal("Run failed")
	}
}

func TestRunServers(t *testing.T) {
	s1 := NewServer("127.0.0.1:")
	s1.Handle("GET", "/", func(http.ResponseWriter, *http.Request) {})

	s2 := NewServer("127.0.0.1:")
	s2.Handle("GET", "/", func(http.ResponseWriter, *http.Request) {})

	// auto close servers
	go func() {
		time.Sleep(time.Millisecond)
		time.Sleep(time.Millisecond)
		s1.Close()
	}()
	go func() {
		time.Sleep(time.Millisecond)
		time.Sleep(time.Millisecond)
		s2.Close()
	}()

	err := RunServers(s1, s2)

	if err != nil && err.Error() != "http: Server closed" {
		t.Fatalf("RunServers failed: %s", err)
	}

	// wait servers to be closed
	time.Sleep(time.Millisecond * 2)
}

func TestRunServersWaitAll(t *testing.T) {
	s1 := NewServer("127.0.0.1:")
	s1.Handle("GET", "/", func(http.ResponseWriter, *http.Request) {})

	s2 := NewServer("127.0.0.1:")
	s2.Handle("GET", "/", func(http.ResponseWriter, *http.Request) {})

	// auto close servers
	go func() {
		time.Sleep(time.Millisecond)
		time.Sleep(time.Millisecond)
		s1.Close()
	}()
	go func() {
		time.Sleep(time.Millisecond)
		time.Sleep(time.Millisecond * 2)
		s2.Close()
	}()

	err := RunServersWaitAll(s1, s2)

	if err != nil && err.Error() != "http: Server closed;http: Server closed" {
		t.Fatalf("RunServersWaitAll failed: %s", err)
	}

	// wait servers to be closed
	time.Sleep(time.Millisecond * 4)
}
