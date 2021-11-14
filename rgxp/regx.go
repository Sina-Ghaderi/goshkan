// Copyright 2021 SNIX LLC sina@snix.ir
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// version 2 as published by the Free Software Foundation.
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

package rgxp

import (
	"context"
	"goshkan/opts"
	"goshkan/rgdb"
	"regexp"
	"strings"
)

// used from ntcp package (TLS or HTTP connections)
var regexComp *regexp.Regexp

func RegexpCompiled() *regexp.Regexp { return regexComp }

func compileReg(rwRegex *string) error {
	tempReg, err := regexp.Compile(*rwRegex)
	if err == nil {
		regexComp = tempReg
	}
	return err
}

// load all regex from mysql database and compile them
func LoadRegexpInit(db *rgdb.MariaSQLDB) {
	ctx, cancel := context.WithTimeout(context.Background(), databaseCtxTime)
	defer cancel()
	allReg, err := db.LoadAllRegex(ctx)
	if err != nil {
		opts.MYSQLD(err)
	}

	opts.SYSLOG(readRegex)
	allReg = append(allReg, &rgdb.RegStruct{Regex: regMatchN})

	for i, rg := range allReg {
		allReg[i].Regex = "(" + rg.Regex + ")"
	}
	ppstr := joinRegexTo(allReg, "|")
	if err := compileReg(&ppstr); err != nil {
		opts.OSEXIT(err)
	}
}

// recompile regexp , when new regex string applied by user
func AddRecompileReg(new *string) error {
	regtext := regexComp.String() + "|(" + *new + ")"
	return compileReg(&regtext)
}

func DelRecompileReg(ptrn *string) error {
	cutd := strings.TrimPrefix(
		strings.TrimSuffix(strings.ReplaceAll(
			strings.ReplaceAll(regexComp.String(),
				"("+*ptrn+")", ""), "||", "|"), "|"), "|")
	return compileReg(&cutd)
}

func joinRegexTo(elems rgdb.RegexList, sep string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return elems[0].Regex
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(elems[i].Regex)
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(elems[0].Regex)
	for _, s := range elems[1:] {
		b.WriteString(sep)
		b.WriteString(s.Regex)
	}
	return b.String()
}
