.PHONY: all generate deps output build clean 


COMMON_SELF_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
ROOT_DIR := $(abspath $(shell cd $(COMMON_SELF_DIR)/ && pwd -P))
OUTPUT_DIR := ${ROOT_DIR}/_output

.DEFAULT_GOAL := all

all: build

generate: deps
	${ROOT_DIR}/hack/update-codegen.sh

deps:
	go mod tidy
	go mod vendor
	
output:
	mkdir -p ${OUTPUT_DIR}

build: output generate
	go build -o ${OUTPUT_DIR}/bin/ttl-controller cmd/ttl-controller/main.go

clean:
	rm -rf ${OUTPUT_DIR}