# goshkan
Transparent TLS and HTTP proxy serve &amp; operating on all 65535 ports, with domain regex whitelist and rest api control

feature: 
- tls and http on same port (peyload inspection)
- handle connections with low memory footprint
- regex domain filtering 
- oprate on all ports (with iptables redirect)
- rest api for add/delete domains
- DNAT friendly, find client actual dst port from conntrack table





