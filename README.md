# Go WebExec

Sometimes for conveniance you need to upload a file somewhere like in a container and run remote commmands
where you don't have access to the container runtime API (like in [Gitlab Services](https://docs.gitlab.com/13.10/ee/ci/docker/using_docker_images.html#what-services-are-not-for))

This package provides a dummy endpoint to upload a file on dest, run arbitrary command or upload and run a script.

Also the service could start a service in separate process.

/!\ NEVER USE IT ON PRODUCTION SERVER OR SENSITIVE HOSTS /!\

Please keep in mind that this package had 0 security and anyone could upload and run code on your service.

Use it in an ephemeral and development environment (like a container in a CI).

## Installation

Via Go

```
go get github.com/ahmet2mir/go-webexec
```

## Usage

### Start server and service

```
$ go-webexec -h

Usage of ./go-webexec:
  -args string
        Command args, put all args separated with a space inside a single string.
  -basedir string
        Directory where files are uploaded. (default "/tmp/base-ci")
  -command string
        Command to run in separate goroutine.
  -host string
        Bind address. (default "127.0.0.1")
  -loglevel string
        logging level (default "info")
  -port int
        Bind port. (default 8080)
```

Create a service to run (optional could be ommited)

```bash
$ cat>/tmp/test.sh<<EOF
#!/bin/bash

echo "Starting process with args \$@"
while true; do
  echo .
  sleep 1
done

exit 0
EOF

$ chmod +x /tmp/test.sh
```

Start Server on port 4444 and start the 2nd service

```bash
$ go-webexec -port 4444 -loglevel debug -basedir /tmp/base-ci -command /tmp/test.sh -args "arg1 arg2"
INFO[0000] Running command '/tmp/test.sh' and args 'arg1 arg2' 
INFO[0000] Running HTTP server and listen on: 127.0.0.1:4444 
Starting process with args arg1 arg2
```

### Upload and run a script

HTTP server is started and the command too, now upload and exec an arbitrary script.

Note that the script will be chown to `0755` automatically.

```
$ cat>/tmp/runit.sh<<EOF
#!/bin/bash

echo "Run it with args \$@"
uname -a
exit 0
EOF
```

Upload and run

```
$ curl -XPOST localhost:4444/upload -F 'file=@/tmp/runit.sh' -F exec=true -F args=hello -F timeout=50s
Successfully Uploaded File to /tmp/base-ci/runit.sh
Run it with args hello
Linux go.webexec.dev 4.19.0-6-amd64 #1 SMP Debian 4.19.67-2+deb10u1 (2020-03-23) x86_64 GNU/Linux
```

On server side you will see

```
DEBU[0164] File Name: runit.sh                          
DEBU[0164] File Size: 56                                
DEBU[0164] File Mode: -rwxr-xr-x                        
DEBU[0164] File MIME: map[Content-Disposition:[form-data; name="file"; filename="runit.sh"] Content-Type:[application/octet-stream]] 
INFO[0164] saveFile(): Destination                       fileDest=/tmp/base-ci/runit.sh
INFO[0164] execCommand(): Script                         script=/tmp/base-ci/runit.sh
INFO[0164] execCommand(): Args                           args=hello
INFO[0164] execCommand(): Timeout Value                  timeout=50s
```

### Upload only


If you just need to upload a file and change default mode to 0644

```
$ curl -XPOST localhost:4444/upload -F 'file=@/tmp/runit.sh' -F mode=0644
Successfully Uploaded File to /tmp/base-ci/runit.sh
```

### Exec only

If you just need to run a command

```
$ curl -XPOST localhost:4444/exec -d script=/bin/echo -d args="hello" -d timeout=50s
hello
```
