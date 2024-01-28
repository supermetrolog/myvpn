package main

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
	"github.com/supermetrolog/myvpn/internal/helpers/checkerr"
	"github.com/supermetrolog/myvpn/internal/helpers/command"
	"github.com/supermetrolog/myvpn/internal/tun"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	MTU = 1500
)

type Forwarder struct {
	localConn *net.UDPConn
	connCache *cache.Cache
}

func main() {
	iface, err := tun.CreateTun("10.1.1.1", MTU)
	checkerr.CheckErr("create tun iface error", err)

	log.Printf("Назначаем форвардинг для созданного интерфейса: %s\n", iface.Name())

	cmd := fmt.Sprintf("sysctl -w net.ipv4.ip_forward=1")
	out, err := command.RunCommand(cmd)
	if err != nil {
		checkerr.CheckErr(out, err)
	}

	cmd = fmt.Sprintf("iptables -t nat -A POSTROUTING -s 10.1.1.0/24 ! -d 10.1.1.0/24 -m comment --comment 'vpndemo' -j MASQUERADE")
	out, err = command.RunCommand(cmd)
	if err != nil {
		checkerr.CheckErr(out, err)
	}

	cmd = fmt.Sprintf("iptables -A FORWARD -s 10.1.1.0/24 -m state --state RELATED,ESTABLISHED -j ACCEPT")
	out, err = command.RunCommand(cmd)
	if err != nil {
		checkerr.CheckErr(out, err)
	}

	cmd = fmt.Sprintf("iptables -A FORWARD -d 10.1.1.0/24 -j ACCEPT")
	out, err = command.RunCommand(cmd)
	if err != nil {
		checkerr.CheckErr(out, err)
	}

	conn := createConn()
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		checkerr.CheckErr("conn close error", err)
	}(conn)

	forwarder := &Forwarder{localConn: conn, connCache: cache.New(30*time.Minute, 10*time.Minute)}

	go listenClient(forwarder, iface)
	go listenIface(forwarder, iface)

	termSignal := make(chan os.Signal, 1)
	signal.Notify(termSignal, os.Interrupt, syscall.SIGTERM)
	<-termSignal
	fmt.Println("closing")
}

func listenIface(forwarder *Forwarder, iface *water.Interface) {
	buf := make([]byte, MTU)

	for {
		n, err := iface.Read(buf)
		log.Printf("READED FROM INTERFACE %d", n)
		checkerr.CheckErr("read from iface error", err)

		if n == 0 {
			continue
		}

		header, err := ipv4.ParseHeader(buf[:n])
		srcAddr, dstAddr := header.Src.String(), header.Dst.String()

		key := fmt.Sprintf("%v->%v", dstAddr, srcAddr)

		log.Println("key ", key, forwarder.connCache.Items())

		v, ok := forwarder.connCache.Get(key)
		if ok {
			_, err = forwarder.localConn.WriteToUDP(buf[:n], v.(*net.UDPAddr))
			checkerr.CheckErr("write to conn error", err)
		}

		command.WritePacket(buf[:n])
	}
}

func listenClient(forwarder *Forwarder, iface *water.Interface) {
	for {
		buf := make([]byte, MTU)
		n, addr, err := forwarder.localConn.ReadFromUDP(buf)
		log.Printf("READED FROM UDP TUNNEL %d", n)
		checkerr.CheckErr("read from udp failed", err)

		if !waterutil.IsIPv4(buf) {
			continue
		}

		_, err = iface.Write(buf[:n])
		checkerr.CheckErr("iface write error", err)

		log.Printf("UDP addr from conn. IP: %s, Port: %d", addr.IP.String(), addr.Port)

		command.WritePacket(buf[:n])

		header, err := ipv4.ParseHeader(buf[:n])
		srcAddr, dstAddr := header.Src.String(), header.Dst.String()

		if srcAddr == "" || dstAddr == "" {
			continue
		}
		key := fmt.Sprintf("%v->%v", srcAddr, dstAddr)
		forwarder.connCache.Set(key, addr, cache.DefaultExpiration)
	}
}

func createConn() *net.UDPConn {
	udpAddr, err := net.ResolveUDPAddr("udp", ":9090")
	checkerr.CheckErr("unable to resolve udp addr", err)

	conn, err := net.ListenUDP("udp", udpAddr)
	checkerr.CheckErr("unable to listen udp", err)

	return conn
}
