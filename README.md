# jumphasher
[![circleci](https://circleci.com/gh/iamthebot/jumphasher/tree/master.svg?style=shield&circle-token=b9d627a2e9b5a29a7675a09876585e1bae06bf65)](https://circleci.com/gh/iamthebot/jumphasher)

JumpCloud Password Hashing Server Challenge

## Description
A simple password hashing server with SSL support. A client can submit many concurrent password hashing requests and lookup hash results after a fixed delay.

## Installing
### Local
Jumpcloud has no external dependencies. Assuming a functioning `GOPATH`, simply run:
```bash
go get -v github.com/jumphasher/common
go get -v github.com/jumphasher/api
go build -o $GOPATH/bin/jumphasher github.com/iamthebot/jumphasher/api
```

### Docker
```bash

```


## Endpoints
| Method | Endpoint    | Parameters                   | Client Payload              | Server Payload                                                                                                                       |
|--------|-------------|------------------------------|-----------------------------|--------------------------------------------------------------------------------------------------------------------------------------|
| `POST` | `/hash`     | N/A                          | A password.<br> Eg; `jumpcloud` | A 32 character job ID. Eg; `fcdff9fc6ec44f059164ec51a756524b`                                                                        |
| `GET`  | `/hash`     | `id` the 32 character job ID | N/A                         | If found, a base 64 encoded hash for the job ID. <br> Eg; `7+jtE9tp16UQHMShH1l0uMlq1JF...`                                                |
| `GET`  | `/stats`    | N/A                          | N/A                         | A JSON structure containing total requests and average request handling time in milliseconds.<br> Eg; `{"total": 14000, "average": "1"}` |
| `GET`  | `/shutdown` | N/A                          | N/A                         | Confirmation that shutdown has commenced                                                                                             |

## Parameters
| Parameter       | Description                                                              | Valid Values                                                                                                                            | Default                           |
|-----------------|--------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------|
| `--sslmode`     | Whether to enable SSL                                                    | `hybrid`: both SSL and plain HTTP <br> `exclusive`: SSL only <br> `disabled`: Plain HTTP only                                           | `hybrid`                          |
| `--port`        | Listening port for plain HTTP connections                                | 1-65535                                                                                                                                 | 80                                |
| `--sslport`     | Listening port for HTTPS connections                                     | 1-65535                                                                                                                                 | 443                               |
| `--sslcert`     | Location of X509 SSL certificate                                         | Valid location of certificate. <br> If one is not available at the given location, a self-signed one will be generated                  | `server.crt`                      |
| `--sslkey`      | Location of SSL private key in PEM format                                | Valid location of private key. <br> If one is not available at the given location, an EC private key will be generated using NIST P-256 | `server.pem`                      |
| `--delay`       | Number of seconds to delay hashing requests before they become available | Positive integers                                                                                                                       | 5                                 |
| `--concurrency` | Target concurrency to use for internal workers and data structures       | 1+                                                                                                                                      | Number of logical cores on system |