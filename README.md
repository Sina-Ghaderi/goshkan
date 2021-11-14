# goshkan
Transparent TLS and HTTP proxy serve &amp; operating on all 65535 ports, with domain regex whitelist and rest api control

- tls and http on same port (peyload inspection)
- handle connections with low memory footprint
- regex domain filtering 
- oprate on all ports (with iptables redirect)
- rest api for add/delete domains
- DNAT friendly, find client actual dst port from conntrack table
- written with golang standard packages (except mysql-driver)

#### compile from source
clone this project, use `git clone https://github.com/Sina-Ghaderi/goshkan.git`  
goshkan written with golang, so you need to install compiler `apt install golang`  
finally run `go build` on project root directory to compile source code.  
FYI: pre-compiled goshkan binary is available at [Releases](https://github.com/Sina-Ghaderi/goshkan/releases)

#### required dependency
first of all, goshkan uses mysql server to store regex patterns, so mysql or mariadb server is 
required. on debian mariadb server can be installed by executing `apt install mariadb-server`  
remember to run `mysql_secure_installation` after installation to secure your sql server.  

now its time to create database, user and tables. in order to do this you need to login to mysql with root user: `mysql -u root`  and on mysql shell run these commands: (you should change username and password)
```
CREATE DATABASE Goshkan;
CREATE USER 'username'@'localhost' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON Goshkan.* TO 'username'@'localhost';

CREATE TABLE Goshkan.regext (
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

#### configuration file
configuration file is based on json, the default configuration file path is `./server-config.json`
for using another file as config, you should specify --config flag: `goshkan --config <path/to/file>`

default config file content:  
```
{
    "MYSQL_PASSWORD": "password",
    "MYSQL_USERNAME": "username",
    "DOMAIN_MEMTTL": 60,
    "MYSQL_DATABASE": "goshkan",
    "MYSQL_ADDRESS": "localhost",
    "CONNECT_TIMEOUT": 10,
    "LISTEN_ADDRESS": "127.0.0.1:8443",
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
`LISTEN_ADDRESS`: tls/http proxy listen address and port (string `addr:port`)  
`HTTPAPI_LISTEN`: http rest api listen address and port (string `addr:port`)  
`LOGS_DEBUGGING`: debugging enable (boolean `true|false`)  
`DOMAIN_MEMTTL`: in memory domain cache aging time in second, value 0 disable this 
functionality (integer >= 0)  

summary about  `DOMAIN_MEMTTL`:   
goshkan uses in memory cache (hashtable) to store recently connected domains and addresses, the reason for this is to reduce time complexity.  
in nutshell time complexity is the amount of time taken by an algorithm to run, which in this case is regex matching algorithm. when client connect to upstream host, goshkan store matched upstream address or domain in memory, so for next upcoming connection doesn't have to go through regex matching algorithm, instead uses hashtable with time complexity of O(1). 

domains and addresses would be stored in memory with a timer, when this timer elapsed, domain or address will be removed from memory (age-out)  unless new connection with this domain/address would be established. in this case, the timer will be reset. `DOMAIN_MEMTTL` value indicate this timer time duration (in second). if `DOMAIN_MEMTTL` is 0 memory cache functionality would be disabled entirely.
 
