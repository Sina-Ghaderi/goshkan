// Copyright 2021 SNIX LLC sina@snix.ir
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// version 2 as published by the Free Software Foundation.
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

package main

import (
	"goshkan/api"
	"goshkan/database"
	"goshkan/ntcp"
	"goshkan/options"
	"goshkan/rgxp"
)

func main() {
	options.OptsInitService()         //logging and configs reader
	service := database.NewDatabase() // connection to mysql database
	rgxp.LoadRegexpInit(service)      // read all regex pattens from database
	ntcp.NewProxy().RunProxy()        // tls and http proxy service
	api.NewApid(service).Run()        // run rest api service
}
