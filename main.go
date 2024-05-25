package main

import (
	s "server"
)

func main() {
	s := new(s.Server)
	s.Run()
}
