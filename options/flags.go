package options

import (
	"fmt"
	"os"
)

func flagUsage() {
	fmt.Printf(`usage of goshkan server:
%v -config <server-config-file>	

options:

  --config <file>   path to config file for goshkan server (in json)
                    default path for this file is server-config.json


Copyright (c) 2021 snix.ir, All rights reserved.
Developed BY <Sina Ghaderi> sina@snix.ir
This work is licensed under the terms of GNU General Public license.
Github: github.com/sina-ghaderi and Source: git.snix.ir
`, os.Args[0])
}
