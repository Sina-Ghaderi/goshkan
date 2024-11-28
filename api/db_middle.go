// Copyright 2021 SNIX LLC sina@snix.ir
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// version 2 as published by the Free Software Foundation.
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

package api

import (
	"goshkan/database"
)

func (pr *apiMiddleRegex) errorOrCancel(err error) {
	select {
	case pr.err <- err:
	case <-pr.ctx.Done():
	}
}

func (pr *apiMiddleRegex) sendOnChannel(data interface{}) {
	select {
	case pr.dch <- data:
	case <-pr.ctx.Done():
	}
}

func (pr *apiMiddleRegex) loadAllRegexCTX(h *database.MariaSQLDB) {
	data, err := h.LoadAllRegex(pr.ctx)
	if err != nil {
		pr.errorOrCancel(err)
		return
	}
	pr.sendOnChannel(data)

}

func (pr *apiMiddleRegex) saveRegexToDB(h *database.MariaSQLDB, ptrn *string) {
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

func (pr *apiMiddleRegex) deltRegexFromDB(h *database.MariaSQLDB, regid uint) {
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
