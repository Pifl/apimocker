# apimocker

## Goal

A standalone application that can host simple tcp/http/https mocks. With an emphsis on templating and association for response generation 

## Key Points

- API-First design 
- Associated a response to a feature of the request for matching
- Team separation, one client host multiple mocks which are manage by separate teams without interference 

## Current Features

## Planned Features

- TCP/HTTP mock creation on demand
- Persistance of mocks (local file store)
- Scripting of data objects 
- Templating response objects


## Go Packages

- **main** Handles configuration and start-up
- **api** The exposed api for interfacing with the application
- **host** Underlying platform for hosting mocks e.g. a mock server
- **mock** Mocks which can be deployed on a host
- **access** Handles user accesses (not security just convinience)

## Run Instructions

`git clone https://github.com/Pifl/apimocker.git`
`go build`
`.\apimocker.exe`
