#!/bin/bash

protoc --go_out=. ./greet/greetpb/greet.proto #generate messages
PATH="${PATH}:${HOME}/go/bin" #update your PATH so that the protoc compiler can find the plugins
protoc --go-grpc_out=require_unimplemented_servers=false:. ./greet/greetpb/greet.proto #generate service

protoc --go_out=. ./calculator/calculatorpb/calculator.proto 
PATH="${PATH}:${HOME}/go/bin"
protoc --go-grpc_out=require_unimplemented_servers=false:. ./calculator/calculatorpb/calculator.proto