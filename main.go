package main

import (
	"github.com/2345tech/apollo-agent/apollo"
	"github.com/2345tech/apollo-agent/boot"
	"log"
)

func main() {
	agent := boot.New(
		boot.WithLauncher(boot.NewLog()),
		boot.WithLauncher(boot.NewProfile()),
		boot.WithLauncher(boot.NewSignal()),
	)

	agent.Init().RegisterHandler(apollo.NewHandler())

	if err := agent.Start(); err != nil {
		log.Println(err.Error())
		panic("[PANIC] agent Start failed")
	}
}
