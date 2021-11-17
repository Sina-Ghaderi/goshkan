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
	"goshkan/rgdb"
)

func (pr *apiMiddRegex) errorOrCancel(err error) {
	select {
	case pr.err <- err:
	case <-pr.ctx.Done():
	}
}

func (pr *apiMiddRegex) sendOnChannel(data interface{}) {
	select {
	case pr.dch <- data:
	case <-pr.ctx.Done():
	}
}

func (pr *apiMiddRegex) loadAllRegexCTX(h *rgdb.MariaSQLDB) {
	data, err := h.LoadAllRegex(pr.ctx)
	if err != nil {
		pr.errorOrCancel(err)
		return
	}
	pr.sendOnChannel(data)

}

func (pr *apiMiddRegex) saveRegexToDB(h *rgdb.MariaSQLDB, ptrn *string) {
	if err := h.RegexIFExist(pr.ctx, ptrn); err != nil {
		pr.errorOrCancel(err)
		return
	}

	if err := h.AddNewRegex(pr.ctx, ptrn); err != nil {
		pr.errorOrCancel(err)
		return
	}

	pr.sendOnChannel(nil)
}

func (pr *apiMiddRegex) deltRegexFromDB(h *rgdb.MariaSQLDB, regid uint) {
	regex, err := h.GetRegexByID(pr.ctx, regid)
	if err != nil {
		pr.errorOrCancel(err)
		return
	}

	if err := h.DeleteRegex(pr.ctx, regid); err != nil {
		pr.errorOrCancel(err)
		return
	}

	pr.sendOnChannel(regex)

}
