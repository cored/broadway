package main

import (
	"github.com/namely/broadway/env"
	"github.com/namely/broadway/server"
	"github.com/namely/broadway/store"
)

func main() {
	s := server.New(store.New())
	s.Init()
	err := s.Run(env.ServerHost)
	if err != nil {
		panic(err)
	}

}
