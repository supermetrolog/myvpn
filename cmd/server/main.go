package main

import (
	"fmt"
	"github.com/songgao/water"
	"github.com/supermetrolog/myvpn/internal/helpers/command"
	"github.com/supermetrolog/myvpn/internal/tun"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func checkErr(message string, e error) {
	if e != nil {
		log.Fatalf(message, e)
	}
}

func main() {
	tunIpAddress := "192.168.1.54"

	iface, err := tun.CreateTun(tunIpAddress)
	checkErr("create tun iface error", err)

	conn := createConn()
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		checkErr("conn close error", err)
	}(conn)

	go listenClient(conn, iface)
	go listenIface(iface, conn)

	termSignal := make(chan os.Signal, 1)
	signal.Notify(termSignal, os.Interrupt, syscall.SIGTERM)
	<-termSignal
	fmt.Println("closing")
}

func listenIface(iface *water.Interface, conn *net.UDPConn) {
	buf := make([]byte, 2000)

	for {
		n, err := iface.Read(buf)
		checkErr("read from iface error", err)

		_, err = conn.Write(buf[:n])
		checkErr("write to conn error", err)

		log.Println("START - incoming packet from INTERFACE")
		command.WritePacket(buf[:n])
		log.Println("DONE - incoming packet from INTERFACE")
	}
}

func listenClient(conn *net.UDPConn, iface *water.Interface) {
	for {
		buf := make([]byte, 2000)
		n, addr, err := conn.ReadFromUDP(buf)
		checkErr("read from udp failed", err)
		_, err = iface.Write(buf[:n])
		checkErr("iface write error", err)

		log.Printf("UDP addr from conn. IP: %s, Port: %d", addr.IP.String(), addr.Port)

		log.Println("START - incoming packet from TUNNEL")
		command.WritePacket(buf[:n])
		log.Println("DONE - incoming packet from TUNNEL")
	}
}

func createConn() *net.UDPConn {
	udpAddr, err := net.ResolveUDPAddr("udp", ":9090")
	checkErr("unable to resolve udp addr", err)

	conn, err := net.ListenUDP("udp", udpAddr)
	checkErr("unable to listen udp", err)

	return conn
}
