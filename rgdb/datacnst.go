// Copyright 2021 SNIX LLC sina@snix.ir
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// version 2 as published by the Free Software Foundation.
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

package rgdb

import "time"

const (
	databaseMaxConn = 128
	databaseMaxLife = 5 * time.Minute
)
const (
	sqlConnect = "%v:%v@tcp(%v)/%v?parseTime=true"
	driverName = "mysql"
)

const (
	sqlSelect = "SELECT regexid, regexstr FROM regext"
	sqlRgByID = "SELECT regexstr from regext where regexid=?"
	sqlDelete = "DELETE FROM regext WHERE regexid=?"
	sqlInsert = "INSERT INTO regext (regexstr) VALUES (?)"
)
