# jumphasher
[![circleci](https://circleci.com/gh/iamthebot/jumphasher/tree/master.svg?style=shield&circle-token=b9d627a2e9b5a29a7675a09876585e1bae06bf65)](https://circleci.com/gh/iamthebot/jumphasher)    [![Docker Repository on Quay](https://quay.io/repository/iamthebot/jumphasher/status "Docker Repository on Quay")](https://quay.io/repository/iamthebot/jumphasher) [![Godoc](https://img.shields.io/badge/godoc-ready-blue.svg)](http://godoc.org/github.com/iamthebot/jumphasher/common)


JumpCloud Password Hashing Server Challenge

## Description
A simple password hashing server with SSL support. A client can submit many concurrent password hashing requests and lookup hash results after a fixed delay.

## Installing
### Local
Jumphasher has no external dependencies. Assuming a functioning `GOPATH`, simply run:
```bash
go get -v github.com/jumphasher/common
go get -v github.com/jumphasher/api
go build -o $GOPATH/bin/jumphasher github.com/iamthebot/jumphasher/api
```

### Docker
Assuming your certificate and key are in `SSL_CERT_FOLDER` and named `server.crt` and `server.pem`...
```bash
docker pull quay.io/iamthebot/jumphasher:latest
docker run -v <SSL cert folder>:/mnt/ssl:Z -p <hosthttpport>:80/tcp -p <hosthttpsport>:443/tcp [--net=<docker network>] [--ip=<custom ip>] quay.io/iamthebot/jumphasher:latest
```
If you don't already have a certificate and key ready, map a host folder where you'd like them to be generated and map it to `/mnt/ssl` on the container. Make sure to append `:Z` to the mapping directive if using a distribution that has SELinux enabled so the contexts are managed properly.

The server will listen on port 80 for HTTP and 443 for HTTPS unless you've mapped the ports.



## Server Parameters
| Parameter       | Description                                                              | Valid Values                                                                                                                            | Default                           |
|-----------------|--------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------|
| `--sslmode`     | Whether to enable SSL                                                    | `hybrid`: both SSL and plain HTTP <br> `exclusive`: SSL only <br> `disabled`: Plain HTTP only                                           | `hybrid`                          |
| `--port`        | Listening port for plain HTTP connections                                | 1-65535                                                                                                                                 | 80                                |
| `--sslport`     | Listening port for HTTPS connections                                     | 1-65535                                                                                                                                 | 443                               |
| `--sslcert`     | Location of X509 SSL certificate                                         | Valid location of certificate. <br> If one is not available at the given location, a self-signed one will be generated                  | `server.crt`                      |
| `--sslkey`      | Location of SSL private key in PEM format                                | Valid location of private key. <br> If one is not available at the given location, an EC private key will be generated using NIST P-256 | `server.pem`                      |
| `--delay`       | Number of seconds to delay hashing requests before they become available | Positive integers                                                                                                                       | 5                                 |
| `--concurrency` | Target concurrency to use for internal workers and data structures       | 1+                                                                                                                                      | Number of logical cores on system |

## Endpoints
| Method | Endpoint    | URI Parameters                   | Client Payload              | Server Payload                                                                                                                       |
|--------|-------------|------------------------------|-----------------------------|--------------------------------------------------------------------------------------------------------------------------------------|
| `POST` | `/hash`     | N/A                          | A password.<br> Eg; `jumpcloud` | A 32 character job ID. Eg; `fcdff9fc6ec44f059164ec51a756524b`                                                                        |
| `GET`  | `/hash`     | `id` the 32 character job ID | N/A                         | If found, a base 64 encoded hash for the job ID. <br> Eg; `7+jtE9tp16UQHMShH1l0uMlq1JF...`                                                |
| `GET`  | `/stats`    | N/A                          | N/A                         | A JSON structure containing total requests and average request handling time in milliseconds.<br> Eg; `{"total": 14000, "average": "1"}` |
| `GET`  | `/shutdown` | N/A                          | N/A                         | Confirmation that shutdown has commenced                                                                                             |

## Tutorial
Here, we'll spin up the server with a 60 second job delay, issue some hashing requests, check some stats, check the resulting hashes, and shut the server down.

First, let's spin up the server:
```bash
<server executable> -port=10000 -sslport=20000 -delay=60
```
You should see something like this:
```
2017/04/07 14:47:24 Server now accepting http connections at port 10000
2017/04/07 14:47:24 Server now accepting https connections at port 20000
```

We're now ready to start issuing requests. Note that since we're using a self-signed certificate in this example, if you're using an automated API testing tool like Postman, you'll have to disable SSL certificate validation. If using `curl`, you'll need to pass `-k` to skip validation against the built-in bundle.

Let's foolishly send a hashing request for a password over  unsecured HTTP:
```bash
curl -w "\n" -X POST -d "hunter2" http://localhost:10000
ce0fc009504f498b5cb42f0b29735c30
```
You'll see a different job ID, but you get the idea. The Job ID is actually a v4 UUID, guaranteed to have an extremely low collision probability.

Now let's be a bit more sensible and issue a hashing request over HTTPS:
```bash
curl -w "\n" -X POST -d "hunter2" https://localhost:20000
d4b49ca1e3f64f206339a12d0307fdf3
```
Since we're now communicating via a secure connection, and we use a cryptographically secure entropy source for our UUIDs, the job ID can now be thought of as an authentication token to retrieve our response. It has 122 bits of entropy which is plenty if were were to implement rate-limiting (not implemented). A trivial implementation of rate limiting would just use a concurrent token bucket algorithm.


OK, now let's go ahead and try to fetch that last one. Again, change the `jobid` to whatever you received from the request.

```
 curl -w "\n" -k https://localhost:20000/hash?id=d4b49ca1e3f64f206339a12d0307fdf3
a5ftaNFOs/GqlZzl1Jx9xhLh6x2v1zsecFhHSD/WpsgJ8s606N9v+ZhMYpj/AoXKzmYUv42qnwBwEBtsiYmeIg==
```

If 60 seconds haven't yet elapsed or you've entered it in wrong, you'll instead get a 404 status code and see something like:
```
hash for job id d4b49ca1e3f64f206339a12d0307fdf3 not found
```
Now, let's check the server stats:
```
curl -w "\n" -k https://localhost:20000/stats
{"total":2,"average":0}
```
Indeed, we've sent two requests. The average is unsurprising since the server isn't under any kind of load, so requests should take under 1 millisecond.

Alright, we've had enough fun. Let's shut the server down.
```
curl -w "\n" -k https://localhost:20000/shutdown
commencing shutdown
```
You'll see something like the following printed to stdout:
```
2017/04/07 15:16:19 Received shutdown request. Commencing shutdown
2017/04/07 15:16:19 Waiting for workers to finish...
```

Try it again while there's a hashing job pending. You'll get a 400 status code.
```
curl -w "\n" -k https://localhost:20000/shutdown
server already shutting down
```

The server should now exit gracefully after roughly `delay` seconds since the last hash request.
