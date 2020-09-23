package main

import (
    _ "fmt"
	"api"
	_ "log"
	_ "host"
	_ "mock"
	_ "access"
)

func main() {
	api.Start(":5050")	
}