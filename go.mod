module github.com/pifl/apimocker

go 1.15

replace api => /api

replace host => /host

replace mock => /mock

replace access => /access

require (
	access v0.0.0-00010101000000-000000000000
	api v0.0.0-00010101000000-000000000000
	github.com/gorilla/mux v1.8.0 // indirect
	host v0.0.0-00010101000000-000000000000
	mock v0.0.0-00010101000000-000000000000
)
