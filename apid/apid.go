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
	"context"
	"encoding/json"
	"goshkan/ntcp"
	"goshkan/opts"
	"goshkan/rgdb"
	"goshkan/rgxp"
	"net/http"
	"regexp"
)

type ApiHandler struct {
	database *rgdb.MariaSQLDB
}

type apiMiddRegex struct {
	ctx context.Context
	err chan error
	dch chan interface{}
}

func newapiListRegex(c context.Context) *apiMiddRegex {
	var err, dch = make(chan error), make(chan interface{})
	return &apiMiddRegex{ctx: c, err: err, dch: dch}
}

func NewApid(sqlhandler *rgdb.MariaSQLDB) *ApiHandler {
	return &ApiHandler{database: sqlhandler}
}

func newHTTPServer(router *http.ServeMux) *http.Server {
	return &http.Server{
		Addr:           opts.Settings.ApiAdd,
		Handler:        router,
		ReadTimeout:    writeTime,
		WriteTimeout:   readxTime,
		MaxHeaderBytes: maxHXSize,
	}
}

func (pr *ApiHandler) Run() {
	router := http.NewServeMux()
	router.HandleFunc(rootShoPDF, pr.rootHandleFX)
	router.HandleFunc(urlListAll, pr.listAllRegex)
	router.HandleFunc(addNewREGX, pr.addNewRegexp)
	router.HandleFunc(deleteREGX, pr.deleteRegexp)

	opts.APISRV(startAPID, opts.Settings.ApiAdd)

	defer pr.database.Handler.Close()
	opts.APIERR(newHTTPServer(router).ListenAndServe())

}

func (pr *ApiHandler) rootHandleFX(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != rootShoPDF {
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
	opts.APIDBG(dbugNewRq, r.RemoteAddr, r.RequestURI)
	if r.URL.Path != urlListAll {
		httpError(w, r, http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodGet {
		httpError(w, r, http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbMiddleCTX)
	defer cancel()

	middle := newapiListRegex(ctx)
	go middle.loadAllRegexCTX(pr.database)

	select {
	case err := <-middle.err:
		opts.APIDBG(err, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))

	case dba := <-middle.dch:
		httpOKLog(w, r, dba.(rgdb.RegexList))

	case <-r.Context().Done():
		opts.APIDBG(ctxIsDone, r.RemoteAddr, r.RequestURI)
	}

}

func (pr *ApiHandler) addNewRegexp(w http.ResponseWriter, r *http.Request) {
	opts.APIDBG(dbugNewRq, r.RemoteAddr, r.RequestURI)
	if r.URL.Path != addNewREGX {
		httpError(w, r, http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		httpError(w, r, http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxBDSise)
	var input rgdb.RegStruct
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&input); err != nil {
		opts.APIDBG(err, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if len(input.Regex) < 1 {
		opts.APIDBG(errRgISZero, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errRgISZero))
		return
	}

	if _, err := regexp.Compile(input.Regex); err != nil {
		opts.APIDBG(err, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbMiddleCTX)
	defer cancel()

	middle := newapiListRegex(ctx)
	go middle.saveRegexToDB(pr.database, input.Regex)

	select {
	case err := <-middle.err:
		opts.APIDBG(err, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))

	case <-middle.dch:
		rgxp.AddRecompileReg(&input.Regex)
		w.WriteHeader(http.StatusOK)

	case <-r.Context().Done():
		opts.APIDBG(ctxIsDone, r.RemoteAddr, r.RequestURI)
	}
}

func (pr *ApiHandler) deleteRegexp(w http.ResponseWriter, r *http.Request) {
	opts.APIDBG(dbugNewRq, r.RemoteAddr, r.RequestURI)
	if r.URL.Path != deleteREGX {
		httpError(w, r, http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		httpError(w, r, http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxBDSise)
	var input rgdb.RegStruct
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&input); err != nil {
		opts.APIDBG(err, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if input.RegID == 0 {
		opts.APIDBG(errIdIsZero, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errIdIsZero))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbMiddleCTX)
	defer cancel()

	middle := newapiListRegex(ctx)
	go middle.deltRegexFromDB(pr.database, input.RegID)

	select {
	case err := <-middle.err:
		opts.APIDBG(err, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))

	case regs := <-middle.dch:
		deltRegxOK(w, r, regs.(*string))

	case <-r.Context().Done():
		opts.APIDBG(ctxIsDone, r.RemoteAddr, r.RequestURI)
	}

}

func deltRegxOK(w http.ResponseWriter, r *http.Request, delreg *string) {
	if err := rgxp.DelRecompileReg(delreg); err != nil {
		opts.APIDBG(err, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if err := ntcp.RemoveFromMap(*delreg); err != nil {
		opts.APIDBG(err, r.RemoteAddr, r.RequestURI)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func httpError(w http.ResponseWriter, r *http.Request, code int) {
	http.Error(w, http.StatusText(code), code)
	opts.APIDBG(http.StatusText(code), r.RemoteAddr, r.RequestURI)
}

func httpOKLog(w http.ResponseWriter, r *http.Request, d rgdb.RegexList) {
	enc := json.NewEncoder(w)
	enc.SetIndent(jsonDefault, jsonNEWPRTX)
	w.WriteHeader(http.StatusOK)
	if err := enc.Encode(d); err != nil {
		opts.APIDBG(err, r.RemoteAddr, r.RequestURI)
	}
}
