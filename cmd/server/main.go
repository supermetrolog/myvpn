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
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Forwarder struct {
	localConn *net.UDPConn
	connCache *cache.Cache
}

func main() {
	tunIpAddress := "192.168.1.54"

	iface, err := tun.CreateTun(tunIpAddress)
	checkerr.CheckErr("create tun iface error", err)

	conn := createConn()
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		checkerr.CheckErr("conn close error", err)
	}(conn)

	forwarder := &Forwarder{localConn: conn, connCache: cache.New(30*time.Minute, 10*time.Minute)}

	go runTestServer("192.168.1.54")
	go listenClient(forwarder, conn, iface)
	go listenIface(forwarder, iface, conn)

	termSignal := make(chan os.Signal, 1)
	signal.Notify(termSignal, os.Interrupt, syscall.SIGTERM)
	<-termSignal
	fmt.Println("closing")
}

func runTestServer(ip string) {
	http.HandleFunc("/hi", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(fmt.Sprintf("hi %s\n", request.RemoteAddr)))
		return
	})
	err := http.ListenAndServe(fmt.Sprintf("%s:8080", ip), nil)
	if err != nil {
		log.Println(err)
	}
}

func listenIface(forwarder *Forwarder, iface *water.Interface, conn *net.UDPConn) {
	buf := make([]byte, 2000)

	for {
		n, err := iface.Read(buf)
		log.Printf("Readed from iface %d", n)
		checkerr.CheckErr("read from iface error", err)

		if n == 0 {
			continue
		}

		//_, err = conn.Write(buf[:n])
		//checkerr.CheckErr("write to conn error", err)

		header, err := ipv4.ParseHeader(buf[:n])
		srcAddr, dstAddr := header.Src.String(), header.Dst.String()

		key := fmt.Sprintf("%v->%v", dstAddr, srcAddr)
		log.Println("key ", key, forwarder.connCache.Items())
		v, ok := forwarder.connCache.Get(key)
		if ok {
			// encrypt data
			_, err = forwarder.localConn.WriteToUDP(buf[:n], v.(*net.UDPAddr))
			checkerr.CheckErr("write to conn error", err)
		}

		log.Println("START - incoming packet from INTERFACE")
		command.WritePacket(buf[:n])
		log.Println("DONE - incoming packet from INTERFACE")
	}
}

func listenClient(forwarder *Forwarder, conn *net.UDPConn, iface *water.Interface) {
	for {
		buf := make([]byte, 2000)
		n, addr, err := conn.ReadFromUDP(buf)
		checkerr.CheckErr("read from udp failed", err)

		if !waterutil.IsIPv4(buf) {
			continue
		}

		_, err = iface.Write(buf[:n])
		checkerr.CheckErr("iface write error", err)

		log.Printf("UDP addr from conn. IP: %s, Port: %d", addr.IP.String(), addr.Port)

		log.Println("START - incoming packet from TUNNEL")
		command.WritePacket(buf[:n])
		log.Println("DONE - incoming packet from TUNNEL")

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
