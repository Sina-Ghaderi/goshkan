// Copyright 2021 SNIX LLC sina@snix.ir
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// version 2 as published by the Free Software Foundation.
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

package opts

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
)

type ServiceSetting struct {
	DbPass string `json:"MYSQL_PASSWORD"`
	DbUser string `json:"MYSQL_USERNAME"`
	Clearc uint   `json:"DOMAIN_MEMTTL"`
	DbName string `json:"MYSQL_DATABASE"`
	DbAddr string `json:"MYSQL_ADDRESS"`
	Conout uint32 `json:"CONNECT_TIMEOUT"`
	SockAd string `json:"LISTEN_ADDRESS"`
	ToutCl uint32 `json:"CLIENT_TIMEOUT"`
	ApiAdd string `json:"HTTPAPI_LISTEN"`
	SDebug bool   `json:"LOGS_DEBUGGING"`
}

var Settings ServiceSetting

func OptsInitService() {
	l := newLogger()
	l.initOptsLogs()
	getAllOptsFile()

	// if debugging is true, enable it
	if Settings.SDebug {
		l.debugging()
	}
}

func getAllOptsFile() {
	flag.Usage = flagUsage
	syspath := flag.String("config", "server-config.json", "config file for go-shkan proxy server")
	flag.Parse()
	file, err := os.Open(*syspath)
	if err != nil {
		OSEXIT(err)
	}
	defer file.Close()
	var content = new(bytes.Buffer)
	_, err = content.ReadFrom(file)
	if err != nil {
		CONFIG(err)
	}

	dec := json.NewDecoder(content)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&Settings); err != nil {
		CONFIG(err)
	}
	if err := checkValueOfConfigType(); err != nil {
		CONFIG(err)
	}

	if Settings.ApiAdd == Settings.SockAd {
		CONFIG(sameAddr)
	}
}

func checkValueOfConfigType() error {
	v := reflect.ValueOf(Settings)
	for i := 0; i < v.NumField(); i++ {
		switch v.Field(i).Kind() {
		case reflect.String:
			in, _ := v.Field(i).Interface().(string)
			if len(in) == 0 {
				return fmt.Errorf(errNoVal, v.Type().Field(i).Name)
			}
		case reflect.Uint32:
			in, _ := v.Field(i).Interface().(uint32)

			if in == 0 && v.Type().Field(i).Name != memcac {
				return fmt.Errorf(errZero, v.Type().Field(i).Name)
			}
		}
	}
	return nil
}
