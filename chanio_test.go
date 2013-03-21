// Copyright 2013 Andreas Louca. All rights reserved.
// Use of this source code is goverend by a BSD-style
// license that can be found in the LICENSE file.

package ChanIO

import (
	"net"
	"testing"
	"time"
)

type SamplePing struct {
	Message string
}

func TestRX(t *testing.T) {
	t.Log("Starting server")
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:8901")
	if err != nil {
		t.Fatalf("Unable to listen: %s\n", err.Error())
	}
	t.Log("Started net listen\n")

	go func() {
		t.Log("Starting server subroutine\n")
		sconn, err := ln.Accept()
		if err != nil {
			t.Fatalf("Unable to accept: %s\n", err.Error())
		}

		t.Log("Accepted connection")

		scio := NewChanIO(sconn)
		scio.SetDebug(true)
		scio.SetVerbose(true)
		srx := scio.GetReceiveChan()
		stx := scio.GetTransmitChan()
		// Respond to Ping
		for {
			t.Log("Waiting for message")
			select {
			case m := <-srx:
				t.Logf("Received message from client %v\n", m)
				// Ping it back
				stx <- m
			}
		}
	}()

	go func() {

		conn, err := net.Dial("tcp", "127.0.0.1:8901")
		if err != nil {
			t.Fatalf("Unable to connect: %s\n", err.Error())
		}

		t.Log("Connected to ChanIO server")
		ccio := NewChanIO(conn)
		if ccio == nil {
			t.Fatal("Invalid ChanIO Endpoint")
		}
		ccio.SetDebug(true)
		ccio.SetVerbose(true)
		crx := ccio.GetReceiveChan()
		ctx := ccio.GetTransmitChan()
		t.Log("Preparing to sent message")

		ctx <- &SamplePing{"Testing!"}

		t.Log("Sent message")

		//time.Sleep(time.Minute)

		// Wait for a message from the network & quit
		for {
			select {
			case m := <-crx:
				t.Logf("Received message from server: %v\n", m)
				return //return
			}
		}

	}()
	time.Sleep(time.Second)
}
