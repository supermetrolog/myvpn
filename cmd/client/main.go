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

const (
	MTU = 1330
)

func main() {
	iface, err := tun.CreateTun("10.1.1.5", MTU)
	checkerr.CheckErr("create tun iface error", err)

	log.Printf("Назначаем форвардинг для созданного интерфейса: %s\n", iface.Name())

	cmd := fmt.Sprintf("sysctl -w net.ipv4.ip_forward=1")
	out, err := command.RunCommand(cmd)
	if err != nil {
		checkerr.CheckErr(out, err)
	}

	cmd = fmt.Sprintf("iptables -t nat -A POSTROUTING -o tun0 -j MASQUERADE")
	out, err = command.RunCommand(cmd)
	if err != nil {
		checkerr.CheckErr(out, err)
	}

	cmd = fmt.Sprintf("iptables -I FORWARD 1 -i tun0 -m state --state RELATED,ESTABLISHED -j ACCEPT")
	out, err = command.RunCommand(cmd)
	if err != nil {
		checkerr.CheckErr(out, err)
	}

	cmd = fmt.Sprintf("iptables -I FORWARD 1 -o tun0 -j ACCEPT")
	out, err = command.RunCommand(cmd)
	if err != nil {
		checkerr.CheckErr(out, err)
	}

	//cmd = fmt.Sprintf("ip route add %s via eth0")
	//out, err = command.RunCommand(cmd)
	//if err != nil {
	//	checkerr.CheckErr(out, err)
	//}

	//cmd = fmt.Sprintf("ip route del 0/1")
	//out, err = command.RunCommand(cmd)
	//if err != nil {
	//	checkerr.CheckErr(out, err)
	//}

	//cmd = fmt.Sprintf("ip route del 128/1")
	//out, err = command.RunCommand(cmd)
	//if err != nil {
	//	checkerr.CheckErr(out, err)
	//}

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
	buf := make([]byte, MTU)

	for {
		n, addr, err := conn.ReadFromUDP(buf)
		log.Printf("READED FROM UDP TUNNEL %d", n)
		checkerr.CheckErr("read from udp error", err)

		_, err = iface.Write(buf[:n])
		checkerr.CheckErr("write to iface error", err)

		log.Printf("UDP addr from conn. IP: %s, Port: %d", addr.IP.String(), addr.Port)

		command.WritePacket(buf[:n])
	}
}

func listenIface(iface *water.Interface, conn *net.UDPConn) {
	buf := make([]byte, MTU)

	for {
		n, err := iface.Read(buf)
		log.Printf("READED FROM INTERFACE %d", n)
		checkerr.CheckErr("read from iface error", err)
		_, err = conn.Write(buf[:n])
		checkerr.CheckErr("write to udp conn error", err)

		command.WritePacket(buf[:n])
	}
}
