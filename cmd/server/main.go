package main

import (
	"github.com/supermetrolog/myvpn/internal/tun"
	"log"
	"time"
)

func main() {

	_, err := tun.CreateTun("192.168.1.54")
	time.Sleep(20 * time.Second)
	log.Fatal(err)
}
