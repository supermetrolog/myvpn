package main

import (
	"fmt"
	"github.com/songgao/water"
	"github.com/supermetrolog/myvpn/internal/helpers/checkerr"
	"github.com/supermetrolog/myvpn/internal/helpers/command"
	"github.com/supermetrolog/myvpn/internal/tun"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	iface, err := tun.CreateTun("192.168.1.12")
	checkerr.CheckErr("create tun iface error", err)

	conn, err := createConn()
	checkerr.CheckErr("create udp conn error", err)

	go listenServer(conn, iface)
	go listenIface(iface, conn)

	termSignal := make(chan os.Signal, 1)
	signal.Notify(termSignal, os.Interrupt, syscall.SIGTERM)
	<-termSignal
	fmt.Println("closing")
}

func createConn() (*net.UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", "server:9090")
	checkerr.CheckErr("resolve udp addr error", err)

	return net.DialUDP("udp", nil, udpAddr)
}

func listenServer(conn *net.UDPConn, iface *water.Interface) {
	buf := make([]byte, 2000)

	for {
		n, addr, err := conn.ReadFromUDP(buf)
		checkerr.CheckErr("read from udp error", err)

		_, err = iface.Write(buf[:n])
		checkerr.CheckErr("write to iface error", err)

		log.Printf("UDP addr from conn. IP: %s, Port: %d", addr.IP.String(), addr.Port)

		log.Println("START - incoming packet from TUNNEL")
		command.WritePacket(buf[:n])
		log.Println("DONE - incoming packet from TUNNEL")
	}
}

func listenIface(iface *water.Interface, conn *net.UDPConn) {
	buf := make([]byte, 2000)

	for {
		n, err := iface.Read(buf)
		checkerr.CheckErr("read from iface error", err)
		_, err = conn.Write(buf[:n])
		checkerr.CheckErr("write to udp conn error", err)

		log.Println("START - incoming packet from INTERFACE")
		command.WritePacket(buf[:n])
		log.Println("DONE - incoming packet from INTERFACE")
	}
}
