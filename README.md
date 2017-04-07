# jumphasher
[![circleci](https://circleci.com/gh/iamthebot/jumphasher/tree/master.svg?style=shield&circle-token=b9d627a2e9b5a29a7675a09876585e1bae06bf65)](https://circleci.com/gh/iamthebot/jumphasher)

JumpCloud Password Hashing Server Challenge

## Description
A simple password hashing server with SSL support. A client can submit many concurrent password hashing requests and lookup hash results after a fixed delay.

## Endpoints
| Method | Endpoint    | Parameters                   | Client Payload              | Server Payload                                                                                                                       |
|--------|-------------|------------------------------|-----------------------------|--------------------------------------------------------------------------------------------------------------------------------------|
| `POST` | `/hash`     | N/A                          | A password.<br> Eg; `jumpcloud` | A 32 character job ID. Eg; `fcdff9fc6ec44f059164ec51a756524b`                                                                        |
| `GET`  | `/hash`     | `id` the 32 character job ID | N/A                         | If found, a base 64 encoded hash for the job ID. <br> Eg; `7+jtE9tp16UQHMShH1l0uMlq1JF...`                                                |
| `GET`  | `/stats`    | N/A                          | N/A                         | A JSON structure containing total requests and average request handling time in milliseconds.<br> Eg; `{"total": 14000, "average": "1"}` |
| `GET`  | `/shutdown` | N/A                          | N/A                         | Confirmation that shutdown has commenced                                                                                             |
