// Copyright 2021 SNIX LLC sina@snix.ir
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// version 2 as published by the Free Software Foundation.
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"goshkan/options"
	"time"
)

const (
	databaseMaxConn = 128
	databaseMaxLife = 5 * time.Minute
)
const (
	sqlConnect = "%v:%v@tcp(%v)/%v?parseTime=true"
	driverName = "mysql"
)

const (
	regexExtin = "err: this pattern already exists in database"
)

const (
	sqlSelect = "SELECT regexid, regexstr FROM regext"
	sqlRgByID = "SELECT regexstr from regext where regexid=?"
	sqlDelete = "DELETE FROM regext WHERE regexid=?"
	sqlExistX = "SELECT EXISTS(SELECT * FROM regext WHERE regexstr=?)"
	sqlInsert = "INSERT INTO regext (regexstr) VALUES (?)"
)

type MariaSQLDB struct {
	Handler *sql.DB
}

type RegStruct struct {
	Regex string
	RegID uint
}

type RegexList []*RegStruct

func NewDatabase() *MariaSQLDB {
	mysqld, err := connectionToDB(
		fmt.Sprintf(sqlConnect, options.Settings.DbUser,
			options.Settings.DbPass, options.Settings.DbAddr, options.Settings.DbName))
	if err != nil {
		options.MYSQLD(err)
	}
	mysqld.Handler.SetMaxIdleConns(databaseMaxConn)
	mysqld.Handler.SetConnMaxLifetime(databaseMaxLife)
	return mysqld
}

func connectionToDB(connString string) (*MariaSQLDB, error) {
	db, err := sql.Open(driverName, connString)
	if err != nil {
		return nil, err
	}
	return &MariaSQLDB{Handler: db}, nil
}

func (d *MariaSQLDB) LoadAllRegex(ctx context.Context) (RegexList, error) {

	sqlRows, err := d.Handler.QueryContext(ctx, sqlSelect)
	if err != nil {
		return nil, err
	}
	defer sqlRows.Close()
	var allRegx RegexList
	for sqlRows.Next() {
		var solreg RegStruct
		if err := sqlRows.Scan(&solreg.RegID, &solreg.Regex); err != nil {
			options.MYSQLO(err)
			continue
		}
		allRegx = append(allRegx, &solreg)
	}
	err = sqlRows.Err()
	return allRegx, err
}

func (d *MariaSQLDB) DeleteRegex(ctx context.Context, rgid uint) error {
	_, err := d.Handler.ExecContext(ctx, sqlDelete, rgid)
	return err
}

func (d *MariaSQLDB) AddNewRegex(ctx context.Context, rgst *string) error {
	_, err := d.Handler.ExecContext(ctx, sqlInsert, *rgst)
	return err
}

func (d *MariaSQLDB) GetRegexByID(ctx context.Context, rgid uint) (*string, error) {
	regex := new(string)
	err := d.Handler.QueryRowContext(ctx, sqlRgByID, rgid).Scan(regex)
	return regex, err
}

func (d *MariaSQLDB) RegexIFExist(ctx context.Context, ptrn *string) error {
	var esd bool
	err := d.Handler.QueryRowContext(ctx, sqlExistX, *ptrn).Scan(&esd)
	if err == nil && esd {
		return errors.New(regexExtin)
	}
	return err
}
