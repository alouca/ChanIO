// Copyright 2012 Andreas Louca. All rights reserved.
// Use of this source code is goverend by a BSD-style
// license that can be found in the LICENSE file.

package ChanIO

import (
	"bufio"
	"encoding/json"
	"github.com/alouca/gologger"
	"io"
)

type ChanIO struct {
	bio *bufio.ReadWriter
	l   *logger.Logger
	rx  chan interface{}
	tx  chan interface{}
}

func NewChanIO(conn io.ReadWriter) *ChanIO {
	cio := new(ChanIO)
	cio.bio = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	cio.l = logger.CreateLogger(false, false)

	// Create the RX/TX Channels
	cio.rx = make(chan interface{}, 100)
	cio.tx = make(chan interface{}, 100)

	go cio.rxWatcher()
	go cio.txWatcher()

	return cio
}

func (c *ChanIO) SetDebug(debug bool) {
	c.l.DebugFlag = debug
}

func (c *ChanIO) SetVerbose(verbose bool) {
	c.l.VerboseFlag = verbose
}

// Returns the Transmit channel interface
func (c *ChanIO) GetTransmitChan() chan<- interface{} {
	var tx chan<- interface{} = c.tx

	return tx
}

// Returns the Receive channel interface
func (c *ChanIO) GetReceiveChan() <-chan interface{} {
	var rx <-chan interface{} = c.rx
	return rx
}

func (c *ChanIO) rxWatcher() {
	c.l.Debug("Starting RX Watcher\n")
	for {
		buf, err := c.bio.ReadBytes(byte(10))
		c.l.Debug("Received data\n")
		if err != nil {
			c.l.Error("Error reading data from channel transport: %s\n", err.Error())
			continue
		}
		var data interface{}
		err = json.Unmarshal(buf, &data)
		if err != nil {
			c.l.Error("Error unmarshalling JSON from channel transport: %s\n", err.Error())
			continue
		}
		c.rx <- data
		c.l.Debug("Read & dispatched %d bytes of data\n", len(buf))
	}
}

func (c *ChanIO) txWatcher() {
	c.l.Debug("Starting TX Watcher\n")
	for {
		select {
		case d := <-c.tx:
			c.l.Debug("Received data for sending\n")
			data, err := json.Marshal(d)
			c.l.Debug("Sending %d bytes\n", len(data))

			if err != nil {
				c.l.Error("Error marshalling data for channel transport: %s\n", err.Error())
				continue
			}
			// Append new line at the end
			data = append(data, byte(10))
			n, err := c.bio.Write(data)
			if err != nil {
				c.l.Error("Error writing data for channel transport: %s\n", err.Error())
				continue
			}
			c.bio.Flush()
			c.l.Debug("Wrote %d bytes to bufio\n", n)
		}
	}
}
