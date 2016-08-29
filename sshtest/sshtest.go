// Package sshtest provides shared test features for SSH functionality.
package sshtest

import (
	"crypto/rand"
	"crypto/rsa"
	"net"
	"strconv"

	"golang.org/x/crypto/ssh"
)

// sshTestConfig returns the options for the SSH test server.
func sshTestConfig() *ssh.ServerConfig {
	var c ssh.ServerConfig
	c.PublicKeyCallback = func(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		// Always allow public key connections, regardless of the key.
		return nil, nil
	}
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		panic(err)
	}

	s, err := ssh.NewSignerFromKey(key)
	if err != nil {
		panic(err)
	}

	c.AddHostKey(s)

	return &c
}

// Server defines the SSH test server, including everything needed to start
// and stop it.
type Server struct {
	// The shutdown channel.
	shutdown chan bool

	// The address of the SSH server (host:port combo).
	Address string
}

// Run starts the server, takes connections to success, and then disconnects
// the client.
func Run() (*Server, error) {
	var s Server
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	s.Address = string(addr.IP) + strconv.Itoa(addr.Port)
	_ = sshTestConfig()

	_, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	go func(s *Server) {
	}(&s)

	return &s, nil
}

// Stop stops the SSH server.
func (s *Server) Stop() error {
	return nil
}
