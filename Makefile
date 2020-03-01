#!/bin/bash

build:
	@echo "${NOW} Building..."
	@go build -o user ./cmd/user/
	@echo "${NOW} Done"

run:
	@make build
	@./user