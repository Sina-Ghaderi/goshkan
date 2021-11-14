// Copyright 2021 SNIX LLC sina@snix.ir
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// version 2 as published by the Free Software Foundation.
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

package apid

import (
	_ "embed"
	"time"
)

const (
	writeTime = 20 * time.Second
	readxTime
)

const (
	dbMiddleCTX = 15 * time.Second
)

const (
	errIdIsZero = "err: RegID cannot be equal or less than zero"
	errRgISZero = "err: Regex cannot be nothing or empty string"
)

const (
	maxHXSize = 1 << 14
	MaxBDSise = 8 << 15
)

const (
	startAPID = "starting api service at:"
	dbugNewRq = "http req:"
	ctxIsDone = "ctx Done() request cancelled by user"
)

const (
	urlListAll = "/api/all/"
	addNewREGX = "/api/add/"
	deleteREGX = "/api/del/"
	rootShoPDF = "/"
)

const (
	jsonDefault = ""
	jsonNEWPRTX = "\x09"
)

//go:embed api.pdf
var embedPDF []byte
