#!/bin/bash

protoc -I rpc/api --go_out=plugins=grpc:rpc/api goget.proto