// Copyright 2021 SNIX LLC sina@snix.ir
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// version 2 as published by the Free Software Foundation.
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

package options

import (
	"log"
	"os"
)

// logging format: prefix after date and time, seems to be fine for me
const logform = log.LstdFlags | log.Lmsgprefix

type logger struct {
	tcplog, conlog *log.Logger
	sysplg, osexit *log.Logger
	config, mysqld *log.Logger
	apisrv         *log.Logger
}

func newLogger() *logger {
	lg := new(logger)
	lg.tcplog = log.New(os.Stdout, "NETLOG:\x20", logform)
	lg.conlog = log.New(os.Stdout, "CONNEC:\x20", logform)
	lg.sysplg = log.New(os.Stdout, "SYSLOG:\x20", logform)
	lg.osexit = log.New(os.Stdout, "OSEXIT:\x20", logform)
	lg.config = log.New(os.Stdout, "CONFIG:\x20", logform)
	lg.mysqld = log.New(os.Stdout, "MYSQLD:\x20", logform)
	lg.apisrv = log.New(os.Stdout, "APISRV:\x20", logform)
	return lg
}

var (
	TLSLOG func(...interface{}) // TLSLOG logs tls server stuff
	SYSLOG func(...interface{})
	OSEXIT func(...interface{}) // EXIT APP
	CONFIG func(...interface{})
	MYSQLD func(...interface{})
	MYSQLO func(...interface{}) // Mysql info
	APISRV func(...interface{})
	APIERR func(...interface{})
	// Debug functions, no-op func non-nil
	APIDBG = func(...interface{}) {}
	CONNEC = func(...interface{}) {}
)

func (lg *logger) initOptsLogs() {
	// new loggers for server handlers, with custom prefix
	TLSLOG = func(l ...interface{}) { lg.tcplog.Println(l...) }
	SYSLOG = func(l ...interface{}) { lg.sysplg.Println(l...) }
	OSEXIT = func(l ...interface{}) { lg.osexit.Fatalln(l...) }
	CONFIG = func(l ...interface{}) { lg.config.Fatalln(l...) }
	MYSQLD = func(l ...interface{}) { lg.mysqld.Fatalln(l...) }
	MYSQLO = func(l ...interface{}) { lg.mysqld.Println(l...) }
	APIERR = func(l ...interface{}) { lg.apisrv.Fatalln(l...) }
	APISRV = func(l ...interface{}) { lg.apisrv.Println(l...) }
}

func (lg *logger) debugging() {
	// do something if debug is true, default is not true
	CONNEC = func(l ...interface{}) { lg.conlog.Println(l...) }
	APIDBG = func(l ...interface{}) { lg.apisrv.Println(l...) }

}
