package tun

import (
	"fmt"
	"github.com/songgao/water"
	"github.com/supermetrolog/myvpn/internal/helpers/command"
	"log"
)

func CreateTun(ip string) (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}

	iface, err := water.New(config)

	if err != nil {
		return nil, fmt.Errorf("create tun iface error: %v", err)
	}

	log.Printf("Created interface with name: %s\n", iface.Name())

	log.Printf("Назначаем IP адресс: %s, для созданного интерфейса: %s\n", ip, iface.Name())

	cmd := fmt.Sprintf("ip addr add %s/24 dev %s", ip, iface.Name())
	//cmd := fmt.Sprintf("sudo ifconfig %s %s netmask 255.255.255.0", iface.Name(), ip)
	out, err := command.RunCommand(cmd)
	if err != nil {
		log.Println(out)
		return nil, err
	}

	log.Println("Включаем созданный интерфейс")

	cmd = fmt.Sprintf("ip link set dev %s up", iface.Name())
	out, err = command.RunCommand(cmd)
	if err != nil {
		log.Println(out)
		return nil, err
	}

	return iface, nil
}
