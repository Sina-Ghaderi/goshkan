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

const (
	domainNotOk = "client trying to connect to blocked domain, killing conn"
	errorChanNk = "cannot terminate domain cleaner task, leaving it in map"
	errorNoHost = "client hostname is unknown cannot do anything here, closing"
)

const (
	startProxyd = "starting tls and http proxy, listening on:"
	disabledMAP = "domain cache time to live is zero, disabling memory cache"
)

const (
	addPortAddr = ":80"
	noneAddress = "none"
	tcpNetConnc = "tcp"
)

const (
	clientConn = "connected to upstream server, address:"
)
