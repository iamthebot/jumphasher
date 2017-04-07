package main

import "github.com/iamthebot/jumphasher/common"

type HashingRequest struct {
	ID         jumphasher.UUID
	Password   []byte
	ReturnChan chan error
}

type HashingResponse struct {
	ID  jumphasher.UUID
	Err error
}
