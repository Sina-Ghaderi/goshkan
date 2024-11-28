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
	"context"
	"encoding/json"

	"goshkan/database"
	"goshkan/ntcp"
	"goshkan/options"
	"goshkan/rgxp"
	"net/http"
	"regexp"
)

type ApiHandler struct {
	database *database.MariaSQLDB
}

type apiMiddleRegex struct {
	ctx context.Context
	err chan error
	dch chan interface{}
}

func newapiListRegex(c context.Context) *apiMiddleRegex {
	var err, dch = make(chan error), make(chan interface{})
	return &apiMiddleRegex{ctx: c, err: err, dch: dch}
}

func NewApid(sqlhandler *database.MariaSQLDB) *ApiHandler {
	return &ApiHandler{database: sqlhandler}
}

func newHTTPServer(router *http.ServeMux) *http.Server {
	return &http.Server{
		Addr:           options.Settings.ApiAddr,
		Handler:        router,
		ReadTimeout:    writeTime,
		WriteTimeout:   readTime,
		MaxHeaderBytes: maxHXSize,
	}
}

func (pr *ApiHandler) Run() {
	router := http.NewServeMux()
	router.HandleFunc(rootSrvPDF, pr.rootHandleFX)
	router.HandleFunc(urlListAll, pr.listAllRegex)
	router.HandleFunc(addNewREGX, pr.addNewRegexp)
	router.HandleFunc(deleteREGX, pr.deleteRegexp)

	options.APISRV(startAPID, options.Settings.ApiAddr)

	defer pr.database.Handler.Close()
	options.APIERR(newHTTPServer(router).ListenAndServe())

}

func (pr *ApiHandler) rootHandleFX(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != rootSrvPDF {
		httpError(w, r, http.StatusBadRequest)
		return
	}
	if r.Method != http.MethodGet {
		httpError(w, r, http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(embedPDF)
}

func (pr *ApiHandler) listAllRegex(w http.ResponseWriter, r *http.Request) {
	options.APIDBG(dbugNewRq, r.RemoteAddr, r.RequestURI)
	if r.URL.Path != urlListAll {
		httpError(w, r, http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodGet {
		httpError(w, r, http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbCTXTimeout)
	defer cancel()

	middle := newapiListRegex(ctx)
	go middle.loadAllRegexCTX(pr.database)

	select {
	case err := <-middle.err:
		options.APIDBG(err, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))

	case dba := <-middle.dch:
		httpOKLog(w, r, dba.(database.RegexList))

	case <-r.Context().Done():
		options.APIDBG(ctxIsDone, r.RemoteAddr, r.RequestURI)
	}

}

func (pr *ApiHandler) addNewRegexp(w http.ResponseWriter, r *http.Request) {
	options.APIDBG(dbugNewRq, r.RemoteAddr, r.RequestURI)
	if r.URL.Path != addNewREGX {
		httpError(w, r, http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		httpError(w, r, http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxBDSise)
	var input database.RegStruct
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&input); err != nil {
		options.APIDBG(err, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if len(input.Regex) < 1 {
		options.APIDBG(errRgISZero, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errRgISZero))
		return
	}

	if _, err := regexp.Compile(input.Regex); err != nil {
		options.APIDBG(err, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbCTXTimeout)
	defer cancel()

	middle := newapiListRegex(ctx)
	go middle.saveRegexToDB(pr.database, &input.Regex)

	select {
	case err := <-middle.err:
		options.APIDBG(err, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))

	case <-middle.dch:
		rgxp.AddRecompileReg(&input.Regex)
		w.WriteHeader(http.StatusOK)

	case <-r.Context().Done():
		options.APIDBG(ctxIsDone, r.RemoteAddr, r.RequestURI)
	}
}

func (pr *ApiHandler) deleteRegexp(w http.ResponseWriter, r *http.Request) {
	options.APIDBG(dbugNewRq, r.RemoteAddr, r.RequestURI)
	if r.URL.Path != deleteREGX {
		httpError(w, r, http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		httpError(w, r, http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxBDSise)
	var input database.RegStruct
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&input); err != nil {
		options.APIDBG(err, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if input.RegID == 0 {
		options.APIDBG(errIdIsZero, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errIdIsZero))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbCTXTimeout)
	defer cancel()

	middle := newapiListRegex(ctx)
	go middle.deltRegexFromDB(pr.database, input.RegID)

	select {
	case err := <-middle.err:
		options.APIDBG(err, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))

	case regs := <-middle.dch:
		deltRegxOK(w, r, regs.(*string))

	case <-r.Context().Done():
		options.APIDBG(ctxIsDone, r.RemoteAddr, r.RequestURI)
	}

}

func deltRegxOK(w http.ResponseWriter, r *http.Request, delreg *string) {
	if err := rgxp.DelRecompileReg(delreg); err != nil {
		options.APIDBG(err, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if err := ntcp.RemoveFromMap(delreg); err != nil {
		options.APIDBG(err, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func httpError(w http.ResponseWriter, r *http.Request, code int) {
	http.Error(w, http.StatusText(code), code)
	options.APIDBG(http.StatusText(code), r.RemoteAddr, r.RequestURI)
}

func httpOKLog(w http.ResponseWriter, r *http.Request, d database.RegexList) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	w.WriteHeader(http.StatusOK)
	if err := enc.Encode(d); err != nil {
		options.APIDBG(err, r.RemoteAddr, r.RequestURI)
	}
}
