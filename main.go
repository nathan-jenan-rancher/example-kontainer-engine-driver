package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/rancher/kontainer-engine/service"
	"github.com/sirupsen/logrus"
)

var wg = &sync.WaitGroup{}

func main() {
	fmt.Println("starting mydriver")
	if os.Args[1] == "" {
		panic(errors.New("no port provided"))
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(fmt.Errorf("argument not parsable as int: %v", err))
	}

	service.RegisterDriverForPort("mydriver", &MyDriver{}, port)

	logrus.Infof("mydriver up and running on port %v", port)

	wg.Add(1)
	wg.Wait() // wait forever, we only exit if killed by parent process
}
