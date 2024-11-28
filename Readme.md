# goshkan
Transparent TLS and HTTP proxy serve &amp; operates on all 65535 ports, with domain regex whitelist and rest api control

- tls and http on same port (payload inspection)
- handle connections with low memory footprint
- regex domain filtering 
- oprate on all ports (with iptables redirect)
- rest api for add/delete domains
- DNAT friendly, find client actual dst port from conntrack table
- written with golang standard packages (except mysql-driver)

### compile from source
clone this project, use `git clone https://github.com/Sina-Ghaderi/goshkan.git`  
mirror snix repository: `gti clone https://git.snix.ir/goshkan`  
goshkan written with golang, so you need to install compiler `apt install golang`  
finally run `go build` on project root directory to compile the source code.  
FYI: pre-compiled goshkan binary is available at [Releases](https://github.com/Sina-Ghaderi/goshkan/releases)

### required dependency
first of all, goshkan uses mysql server to store regex patterns, so mysql or mariadb server is 
required. on debian mariadb server can be installed by executing `apt install mariadb-server`  
remember to run `mysql_secure_installation` after installation to secure your sql server.  

now its time to create database, user and tables. in order to do this you need to login to mysql with root user: `mysql -u root`  and on mysql shell run these commands: (remember to change username and password)
```sql
CREATE DATABASE goshkan;
CREATE USER 'username'@'localhost' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON goshkan.* TO 'username'@'localhost';

CREATE TABLE goshkan.regext (
	regexid INT UNSIGNED auto_increment NOT NULL,
	regexstr LONGTEXT NOT NULL,
	PRIMARY KEY (regexid)
)

ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_general_ci;
FLUSH PRIVILEGES;
EXIT;
```
thats it, mysql installation is completed now. note: if you planning to host mysql and goshkan on separate servers, you should change `localhost` to goshkan server address.

### configuration file
configuration file is based on json, the default configuration file path is `./server-config.json`
for using another file as config, you should specify --config flag: `goshkan --config <path/to/file>`

default config file content:  
```json
{
    "MYSQL_PASSWORD": "password",
    "MYSQL_USERNAME": "username",
    "DOMAIN_MEMTTL": 60,
    "MYSQL_DATABASE": "goshkan",
    "MYSQL_ADDRESS": "localhost",
    "CONNECT_TIMEOUT": 10,
    "LISTEN_ADDRESS": "192.168.122.149:8443",
    "CLIENT_TIMEOUT": 15,
    "HTTPAPI_LISTEN": "127.0.0.1:8080",
    "LOGS_DEBUGGING": true
}
```
`MYSQL_PASSWORD`: database username password (string)  
`MYSQL_USERNAME`: database username (string)  
`MYSQL_DATABASE`: mysql database name (string)  
`MYSQL_ADDRESS`:  mysql server address and port, default port `3306` (string `host:port`)  
`CONNECT_TIMEOUT`: connect to upstream server connection timeout in second (integer > 0)  
`CLIENT_TIMEOUT`: client connection timeout in second (integer > 0)  
`DOMAIN_MEMTTL`: in memory domain cache aging time in second, value 0 disable this 
functionality (integer >= 0)  
`LISTEN_ADDRESS`: tls/http proxy listen address and port (string `addr:port`)  
`HTTPAPI_LISTEN`: http rest api listen address and port (string `addr:port`)  
`LOGS_DEBUGGING`: debugging enable (boolean `true|false`)  


summary about  `DOMAIN_MEMTTL`:   
goshkan uses in memory cache (hashtable) to store recently connected domains and addresses, the reason for this is to reduce time complexity.  
in nutshell time complexity is the amount of time taken by an algorithm to run, which in this case is regex matching algorithm. when client connect to upstream host, goshkan store matched upstream address or domain in memory, so for next upcoming connection doesn't have to go through regex matching again, instead uses hashtable with time complexity of O(1). 

domains and addresses would be stored in memory with a timer, when this timer elapsed, domain or address will be removed from memory (age-out)  unless new connection with this domain/address established before that. in this case, the timer will be reset. `DOMAIN_MEMTTL` value indicate this timer time duration (in second). if `DOMAIN_MEMTTL` set to 0 memory cache functionality would be disabled entirely.  
you should enable it if your server has decent amount of memory (a.k.a RAM)

### iptables redirect all ports
forwarding all ports to goshkan with iptables redirect:  
this command would redirect all tcp packets (on all ports) to goshkan proxy port if packet destination address is 192.168.122.149 and input interface is ens3  

note: after this you can't serve another service on this address (192.168.122.149 in my case) because there is no port left. for solving this issue you may want to exclude your service ports from being forwarding to goshkan proxy port by executing this `iptables -t nat -A PREROUTING -i ens3 -d 192.168.122.149 -p tcp -m tcp --dport 22 -j ACCEPT` command before following one (replace 22 with your service port)  
but the best solution would be to bind your services with another ip-address or interface.

```
iptables -t nat -A PREROUTING -i ens3 -d 192.168.122.149 -p tcp -m tcp --dport 1:65535 -j REDIRECT --to-ports 8443
```

### max open files on linux
by default goshkan can open max 1024 file (connection), if its not enough change this value in systemd service file or with `ulimit` command.  
see [systemd documention](https://www.freedesktop.org/software/systemd/man/systemd.service.html) and [ulimit manual](https://linuxcommand.org/lc3_man_pages/ulimith.html)

### api reference documention

get rest api documention in pdf format by sending `GET /` to `HTTPAPI_LISTEN` address, or [find it](https://github.com/Sina-Ghaderi/goshkan/blob/main/api/api.pdf) under `goshkan/apid/` directory.  
this is open api without authentication, you shouldn't expose it to public, nginx or apache can protect this api with basic http authentication.

### security notice
- do NOT add regex pattern that allows `localhost` , `127.0.0.1` or any of your server ip-address or domains. can cause server connection loop or exposing internal server resources to unauthorized users.

### contribute to this project
feel free to email me <sina@snix.ir> if you want to contribute to this project

Copyright 2021 SNIX LLC sina@snix.ir  
This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License version 2 as published by the Free Software Foundation.  
This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

