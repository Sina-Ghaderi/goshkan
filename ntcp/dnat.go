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
	"fmt"
	"net"
	"strconv"

	"golang.org/x/sys/unix"
)

type sockOpt struct {
	localAddr net.IP
	localPort uint16
}

func newNfTab() *sockOpt { return &sockOpt{localAddr: make(net.IP, 16)} }

func (p *sockOpt) HostString() string {
	return fmt.Sprintf("%v:%d", p.localAddr.To4().String(), p.localPort)
}

func (p *sockOpt) PortNumber() uint16 {
	return p.localPort
}

func (p *sockOpt) PortString() string {
	return strconv.Itoa(int(p.localPort))
}

func (p *sockOpt) IPv4Address() net.IP {
	return p.localAddr.To4()
}

func networkOpt(conn net.Conn) (*sockOpt, error) {
	var nftab = newNfTab()
	fd, err := conn.(*net.TCPConn).File()
	if err != nil {
		return nftab, err
	}

	defer fd.Close()

	raddr, err := unix.GetsockoptIPv6Mreq(
		int(fd.Fd()), unix.IPPROTO_IP, unix.SO_ORIGINAL_DST)
	if err != nil {
		return nftab, err
	}

	nftab.localAddr = net.IP(raddr.Multiaddr[4:8])
	// convert big endian to little endian
	nftab.localPort = uint16(raddr.Multiaddr[2])<<8 + uint16(raddr.Multiaddr[3])
	return nftab, err
}
