// Copyright 2021 SNIX LLC sina@snix.ir
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// version 2 as published by the Free Software Foundation.
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

package ntcp

import (
	"bytes"
	"crypto/tls"
	"errors"
	"goshkan/opts"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

type proxyTLS struct {
	server string
	cntout time.Duration
	readot time.Duration
}

type readOnly struct {
	reader io.Reader
}

func (conn readOnly) Read(p []byte) (int, error)         { return conn.reader.Read(p) }
func (conn readOnly) Write(p []byte) (int, error)        { return 0, io.ErrClosedPipe }
func (conn readOnly) Close() error                       { return nil }
func (conn readOnly) LocalAddr() net.Addr                { return noAddrNetwork{} }
func (conn readOnly) RemoteAddr() net.Addr               { return noAddrNetwork{} }
func (conn readOnly) SetDeadline(t time.Time) error      { return nil }
func (conn readOnly) SetReadDeadline(t time.Time) error  { return nil }
func (conn readOnly) SetWriteDeadline(t time.Time) error { return nil }

type listenHandler struct {
	chnn chan readOnly
}

func (ls listenHandler) Accept() (net.Conn, error) {
	if cox, ok := <-ls.chnn; ok {
		return cox, nil
	}
	return nil, io.ErrClosedPipe
}

func (ls listenHandler) Close() error   { return nil }
func (ls listenHandler) Addr() net.Addr { return noAddrNetwork{} }

type httpImplmnt struct{}

func (httpImplmnt) ServeHTTP(w http.ResponseWriter, r *http.Request) { setHostName(r) }

var setHostName = func(r *http.Request) {}

type noAddrNetwork struct{}

func (noAddrNetwork) Network() string { return noneAddress }
func (noAddrNetwork) String() string  { return noneAddress }

type clientReadHost interface {
	extractHost() (string, error)
	getIoReader() io.Reader
	setIoWriter(io.Writer)
}

type httpHostRead struct{ re, te io.Reader }
type sniTLSLoadHs struct{ re, te io.Reader }

func (rc *httpHostRead) getIoReader() io.Reader   { return rc.re.(*forgeReader).reader }
func (rc *sniTLSLoadHs) getIoReader() io.Reader   { return rc.re.(*forgeReader).reader }
func (rc *httpHostRead) setIoWriter(wr io.Writer) { rc.te = io.TeeReader(rc.re, wr) }
func (rc *sniTLSLoadHs) setIoWriter(wr io.Writer) { rc.te = io.TeeReader(rc.re, wr) }

type forgeReader struct {
	reader io.Reader
	missig io.Reader
}

func (p *forgeReader) Read(b []byte) (int, error) {
	return io.MultiReader(p.missig, p.reader).Read(b)
}

func NewProxy() *proxyTLS {
	sokaddr := opts.Settings.SockAd
	readout := opts.Settings.Conout
	timeout := opts.Settings.ToutCl
	return &proxyTLS{
		server: sokaddr,
		cntout: time.Duration(timeout) * time.Second,
		readot: time.Duration(readout) * time.Second,
	}
}

func (str *proxyTLS) RunProxy() {
	go func() {
		opts.TLSLOG(startProxyd, str.server)
		setupCache()
		soc, err := net.Listen(tcpNetConnc, str.server)
		if err != nil {
			opts.OSEXIT(err)
		}
		defer soc.Close()

		for {
			conn, err := soc.Accept() // accept incomming connections
			if err != nil {
				opts.CONNEC(err)
				continue
			}
			go str.handleNewConn(conn)
		}
	}()
}

func (str *proxyTLS) handleNewConn(conn net.Conn) {
	defer conn.Close()
	// setup an small buffer to capture packet signature
	pktsig := make([]byte, 3)
	if _, err := conn.Read(pktsig); err != nil {
		opts.CONNEC(err)
		return
	}

	// so is this a TLS connection or anything else? 0x16 hello 0x0301 tls version
	// based on tls RFC first packet version is always 0x0301
	if bytes.Equal(pktsig, []byte{0x16, 0x03, 0x01}) {
		str.handleTLSConn(conn, pktsig)
		return
	}

	// Handle HTTP Request
	str.handleHTTPConn(conn, pktsig)
}

func (str *proxyTLS) handleTLSConn(inConn net.Conn, miss []byte) {
	if err := inConn.SetReadDeadline(time.Now().Add(str.cntout)); err != nil {
		opts.CONNEC(err, inConn.RemoteAddr().String())
		return
	}

	header, fread, err := peekClientHost(
		&sniTLSLoadHs{re: &forgeReader{
			reader: inConn, missig: bytes.NewReader(miss)}})
	if err != nil {
		opts.CONNEC(err, inConn.RemoteAddr().String())
		return
	}

	if err := inConn.SetReadDeadline(time.Time{}); err != nil {
		opts.CONNEC(err, inConn.RemoteAddr().String())
		return
	}

	if !allowedOrNot(header) {
		opts.CONNEC(domainNotOk, inConn.RemoteAddr().String())
		return
	}

	sokOpt, err := networkOpt(inConn) // syscall, get port before redirect
	if err != nil {
		opts.CONNEC(err, inConn.RemoteAddr().String())
		return
	}

	dstaddr := net.JoinHostPort(header, sokOpt.PortString())
	outConn, err := net.DialTimeout(tcpNetConnc, dstaddr, str.readot) // outConn
	if err != nil {
		opts.CONNEC(err, inConn.RemoteAddr().String())
		return
	}

	opts.CONNEC(clientConn, dstaddr) // debug logging, if enabled

	storeToMap(header) // cache domain in memory map

	defer outConn.Close()
	// read and write on connections, exchange data...
	readAndwrite(inConn, outConn, fread)

}

func readAndwrite(inConn, outConn net.Conn, buffed io.Reader) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		io.Copy(inConn, outConn) // from dst target to client
		inConn.(*net.TCPConn).CloseWrite()
		wg.Done()
	}()
	go func() {
		io.Copy(outConn, buffed) // from client (with everything readen)
		outConn.(*net.TCPConn).CloseWrite()
		wg.Done()
	}()

	wg.Wait() // wait for them
}

func (str *proxyTLS) handleHTTPConn(inConn net.Conn, miss []byte) {

	if err := inConn.SetReadDeadline(time.Now().Add(str.cntout)); err != nil {
		opts.CONNEC(err, inConn.RemoteAddr().String())
		return
	}

	header, fread, err := peekClientHost(
		&httpHostRead{re: &forgeReader{
			reader: inConn, missig: bytes.NewReader(miss),
		}})
	if err != nil {
		opts.CONNEC(err, inConn.RemoteAddr().String())
		return
	}

	if err := inConn.SetReadDeadline(time.Time{}); err != nil {
		opts.CONNEC(err, inConn.RemoteAddr().String())
		return
	}

	var tmpdtt string
	pindex := whereIsThePort(header, ':')
	switch {
	case pindex < 0:
		header = header + addPortAddr // add port 80 to host
		tmpdtt = header               // for regex check
	default:
		tmpdtt = header[:pindex] // cut port from host, for regex check
	}

	if !allowedOrNot(tmpdtt) {
		opts.CONNEC(domainNotOk, inConn.RemoteAddr().String())
		return
	}

	outConn, err := net.DialTimeout(tcpNetConnc, header, str.readot)
	if err != nil {
		opts.CONNEC(err, inConn.RemoteAddr().String())
		return
	}

	opts.CONNEC(clientConn, header) // debug logging, if enabled

	storeToMap(tmpdtt) // cache in memory
	defer outConn.Close()
	readAndwrite(inConn, outConn, fread)
}

func (pr *httpHostRead) extractHost() (string, error) {
	var finHost string
	var notifychan = make(chan struct{})
	var releaseall = make(chan struct{})
	setHostName = func(r *http.Request) { // change function
		finHost = r.Host         // pass the variable
		notifychan <- struct{}{} // notify others when job is done
	}

	comnuichan := make(chan readOnly) // http serve, accept once.

	// issue: non-http connection stuck on http proxy, bcz protocol is not http
	// connection would be closed before setHostName runs and send notify on notifyc.

	srv := &http.Server{
		Handler: httpImplmnt{},
		ConnState: func(c net.Conn, cs http.ConnState) {
			if cs == 4 { // 4 == connection closed
				select {
				case notifychan <- struct{}{}:
				case <-releaseall:
				}
			}
		},
	}

	go srv.Serve(listenHandler{chnn: comnuichan})

	// put conn on the channel
	comnuichan <- readOnly{reader: pr.te}

	<-notifychan      // job is done
	close(releaseall) // release ConnState function
	close(comnuichan) // close chan to break http serve

	if len(finHost) == 0 { // check finHost
		return finHost, errors.New(errorNoHost)
	}

	return finHost, nil
}

func peekClientHost(rcv clientReadHost) (string, io.Reader, error) {
	peekedBytes := new(bytes.Buffer)
	rcv.setIoWriter(peekedBytes)
	hello, err := rcv.extractHost()
	if err != nil {
		return hello, nil, err
	}
	return hello, io.MultiReader(peekedBytes, rcv.getIoReader()), nil
}

func (pr *sniTLSLoadHs) extractHost() (string, error) {
	var tlsdata string
	err := tls.Server(readOnly{reader: pr.te}, &tls.Config{
		GetConfigForClient: func(tlshand *tls.ClientHelloInfo) (*tls.Config, error) {
			tlsdata = tlshand.ServerName // GetConfigForClient runs after handshake
			return nil, nil
		}}).Handshake() // just handshake

	if len(tlsdata) == 0 {
		return tlsdata, err // check tlsdata
	}
	return tlsdata, nil
}

func whereIsThePort(s string, b byte) int {
	i := len(s)
	for i--; i >= 0; i-- {
		if s[i] == b {
			break
		}
	}
	return i
}
