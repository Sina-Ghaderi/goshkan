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
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
)

type ServiceSetting struct {
	DbPass    string `json:"MYSQL_PASSWORD"`
	DbUser    string `json:"MYSQL_USERNAME"`
	ClearTTL  uint   `json:"DOMAIN_MEMTTL"`
	DbName    string `json:"MYSQL_DATABASE"`
	DbAddr    string `json:"MYSQL_ADDRESS"`
	ConnOut   uint32 `json:"CONNECT_TIMEOUT"`
	SockAddr  string `json:"LISTEN_ADDRESS"`
	InTimeout uint32 `json:"CLIENT_TIMEOUT"`
	ApiAddr   string `json:"HTTPAPI_LISTEN"`
	Debug     bool   `json:"LOGS_DEBUGGING"`
}

var Settings ServiceSetting

func OptsInitService() {
	l := newLogger()
	l.initOptsLogs()
	getAllOptsFile()

	if Settings.Debug {
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

	if Settings.ApiAddr == Settings.SockAddr {
		CONFIG("error: proxy address and http api address cannot be the same")
	}
}

func checkValueOfConfigType() error {
	v := reflect.ValueOf(Settings)
	for i := 0; i < v.NumField(); i++ {
		switch v.Field(i).Kind() {
		case reflect.String:
			in, _ := v.Field(i).Interface().(string)
			if len(in) == 0 {
				return fmt.Errorf(
					"error: field %v has no value, killing main process",
					v.Type().Field(i).Name)
			}
		case reflect.Uint32:
			in, _ := v.Field(i).Interface().(uint32)

			if in == 0 && v.Type().Field(i).Name != "ClearTTL" {
				return fmt.Errorf(
					"error: field %v has no value, killing main process",
					v.Type().Field(i).Name)
			}
		}
	}
	return nil
}
