# Count the ips on your access logs and see the top stuff!

This is probably a very convoluted way of counting and showing the top IP addresses acessing your web server (or servers). It's a three tool... tool?

## ipcount

Runs on your web server, greps the logs and sends the results to Redis. It's a three megabyte binary and takes all the needed config via command line:

Example:
```
/usr/local/bin/ipcount -d 3 -h redisserver:6379 -l /var/log/nginx/access.log -p "" -r '\S+\s+"(\S+)"\s+10\..*'
```

Usage output for the lazy:
```
-d=-1: Redis DB number to store the data
-h="localhost:6379": Hostname:port Redis
-l="": Path to the access log to watch
-p="": Redis password
-r="": PCRE Regex to parse the log. Must return the remote IP on the first capture group
```

I use runit to manage it, but anything will do (it's running for two months without dying at the moment).

## ipcountclean

You only need one instance of this. It will prune the Redis DB and expire IPs that are too old.

Usage:
```
-d=-1: Redis DB number to store the data
-debug=false: Show a lot of probably useless information
-h="localhost:6379": Hostname:port Redis
-p="": Redis password
```

Run it on a cron or similar:

```
*/5 * * * * /usr/local/bin/ipcountclean -d 3 -h localhost:6379
```

## ipcounttop

You also only need one of those. It will read the info from Redis and assemble a (kinda) nice page with tables. You can also click the IP address to get whois info.

You will need:
  - [GeoLite City DB](http://dev.maxmind.com/geoip/legacy/geolite/) It says legacy but still being updated. The city db is needed because geo information.
  - [Mapbox map id](https://www.mapbox.com/) Sign up, create a pretty map and stick it to map.html (the format is `username.randomstring`)
  - The whois command installed.

Example:
```
/usr/local/bin/ipcounttop -d 3 -h redisserver:6379 -s /usr/local/share/ipcounttop -g /usr/share/GeoIP/GeoLiteCity.dat
```

Usage:
```
-d=-1: Redis DB number to store the data
-g="/usr/share/GeoIP/GeoIP.dat": Port to use for webserver :<port number> format
-h="localhost:6379": Hostname:port Redis
-l=":8888": Port to use for webserver :<port number> format
-p="": Redis password
-s="./src/github.com/coredump/ipcount/ipcounttop/webapp": Path to webapp files dir
```

I also use runit for it.

## Blurred screenshot to protect the innocent

![screenshot](http://i.imgur.com/y7Bj4E4.png screenshot)

## BUGS:

It ignores accesses from 127.0.0.1, but to add other IPs to the ignore list you must change code and recompile the binary. Should probably make this easier.

I failed pretty hard on showing the map on a modal, but code still lying around there. I am also deeply embarassed of my javascript coding and will gladly accept push requests to fix/make it better/prettier. Auto refresh would be nice, for example.

I also realize that there are many ways to do the same thing, including most of the log aggregation/analysis tools. BUT someone must have a use for it, so I made it.

## Disclaimers

This product includes GeoLite data created by MaxMind, available from [http:http://www.maxmind.com](http://www.maxmind.com)

Maps are powered by Mapbox and uses a lot of cool open data. [Attribution](https://www.mapbox.com/about/maps/)
