package main

import (
	"fmt"
	"log"
	"os"

	"github.com/namely/broadway/instance"
	"github.com/namely/broadway/playbook"
	"github.com/namely/broadway/server"
	"github.com/namely/broadway/store"
)

func main() {
	/*
		args := os.Args
		yamlFileDescriptor := args[1:][0]
	*/
	playbooks, err := playbook.LoadPlaybookFolder("playbooks/")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v+\n", playbooks)
	fmt.Println(instance.InstanceStatusNew)
	server := server.New(store.New())
	server.Run(os.Getenv("HOST"))

}
