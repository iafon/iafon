package iafon

import (
	"errors"
	"fmt"
	"net/http"
)

type Server struct {
	http.Server

	// only for get methods of Router type
	*Router
}

func NewServer(addr ...string) *Server {
	s := &Server{}

	if len(addr) > 0 {
		s.Addr = addr[0]
	}

	s.Router = newRouter()
	s.Handler = s.Router

	return s
}

func (s *Server) Run() error {
	if s.matcher.Len() == 0 {
		return errors.New("no route added, can not run.")
	}

	fmt.Println("listening on " + s.Addr)

	err := s.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}

	return err
}

func RunServers(servers ...*Server) error {
	return runServers(false, servers...)
}

func RunServersWaitAll(servers ...*Server) error {
	return runServers(true, servers...)
}

func runServers(shouldWaitAll bool, servers ...*Server) error {
	if len(servers) == 0 {
		return errors.New("RunServers require server")
	} else if len(servers) == 1 {
		return servers[0].Run()
	}

	closed := make(chan error, len(servers))

	for _, s := range servers {
		s := s
		go func() {
			err := errors.New("Unkown error")
			defer func() { closed <- err }()
			err = s.Run()
		}()
	}

	if shouldWaitAll {
		return waitAll(closed)
	} else {
		return waitOne(closed)
	}
}

func waitOne(c chan error) error {
	return <-c
}

func waitAll(c chan error) error {
	var err_msg, sep string
	var n = cap(c)

	for {
		select {
		case err := <-c:
			n--
			err_msg += sep + err.Error()
			sep = ";"
		}
		if n <= 0 {
			break
		}
	}

	return errors.New(err_msg)
}
