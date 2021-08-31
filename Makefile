##
# Mastodon Bot
#
# @file
# @version 0.1
include .env
export

build:
	@go build -o bin/main main.go

run:
	@go run main.go


# end
