// Copyright 2021 SNIX LLC sina@snix.ir
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// version 2 as published by the Free Software Foundation.
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

package ntcp

import (
	"errors"
	"goshkan/options"
	"goshkan/rgxp"
	"regexp"
	"sync"
	"time"
)

type memoryTTL struct {
	ticker   *time.Ticker // age-out trigger
	stopfunc func() error // delete by user trigger
}

var inMemMap sync.Map // safe map
var cleanint time.Duration

func setupCache() {
	switch options.Settings.ClearTTL {
	case 0: // zero map ttl, disable map cache
		options.SYSLOG("domain cache time to live is zero, disabling memory cache")
		allowedOrNot = func(domain string) bool { return rgxp.RegexpCompiled().MatchString(domain) }
		storeToMap = func(domain string) {}
		RemoveFromMap = func(*string) error { return nil }

		return
	}

	cleanint = time.Duration(options.Settings.ClearTTL) * time.Second
}

var storeToMap = func(domain string) {
	if ttl, ok := inMemMap.Load(domain); ok {
		ttl.(memoryTTL).ticker.Reset(cleanint) // renew domain age.
		return
	}

	tk := time.NewTicker(cleanint) // new ticker
	ch := make(chan struct{})
	fn := func() error {
		var err error
		select {
		case ch <- struct{}{}:
			return err
		default:
			return errors.New("cannot terminate domain cleaner task, leaving it in map")
		}
	}

	inMemMap.Store(domain, memoryTTL{ticker: tk, stopfunc: fn}) // add data to map
	go cleanerMap(tk, domain, ch)                               // run cleaner
}

func cleanerMap(tk *time.Ticker, dm string, chtr <-chan struct{}) {
	select {
	case <-tk.C: // age out, remove form map
	case <-chtr: // signal by user, remove from map
	}
	tk.Stop() // release ticker, its time to die.
	inMemMap.Delete(dm)

}

var RemoveFromMap = func(ptrn *string) error {
	re, err := regexp.Compile(*ptrn)
	if err != nil {
		return err // somthing is wrong!
	}

	rgtask := func(key interface{}, value interface{}) bool {
		if re.MatchString(key.(string)) {
			if err := value.(memoryTTL).stopfunc(); err != nil {
				options.SYSLOG(err)
			} // remove domain from map
		}
		return true
	}

	inMemMap.Range(rgtask)
	return err
}

var allowedOrNot = func(domain string) bool {
	// check memory cache first -> O(1)
	if _, ok := inMemMap.Load(domain); ok {
		return ok
	}

	// this take time, but we can do nothing about it :/
	return rgxp.RegexpCompiled().MatchString(domain)
}
