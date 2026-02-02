# Simple Http Server for Shell Script

Simple Http Server for Shell Script is an HTTP server based on bash shell scripts.

It is very suitable for quickly developing an HTTP service demo for those who are familiar with and passionate about
bash shell.

Shell script demo: [http-server.sh](http-server.sh)

On Windows, it used `C:/Program Files/Git/bin/bash.exe`, on Linux/macOS, it used `/bin/bash`

## Usage

### command

![](img/01-help-doc.png)

```shell
# show help doc
./http-server -h

# show version
./http-server -v

# run server
./http-server

# use another port
./http-server -p 80

# use another shell
./http-server -s "/path/to/bash"

```

### request

```shell
curl -s -H 'AUTH: xxx' -d '{"auth": "xxx"}' 'http://localhost:8080/users/1?id=1&name=wrj'
```

output:

![](img/02-response.png)
