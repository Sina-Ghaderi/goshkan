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

import "time"

const (
	readRegex = "parsing domain database and compiling regex string patterns"
	regMatchN = "a^" // match nothing, this should be in regex pattern.
)

const (
	databaseCtxTime = 10 * time.Second
)
