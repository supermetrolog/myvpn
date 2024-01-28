package tun

import (
	"fmt"
	"github.com/songgao/water"
	"github.com/supermetrolog/myvpn/internal/helpers/command"
	"log"
)

func CreateTun(ip string, mtu int) (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}

	iface, err := water.New(config)

	if err != nil {
		return nil, fmt.Errorf("create tun iface error: %v", err)
	}

	log.Printf("Created interface with name: %s\n", iface.Name())

	log.Printf("Назначаем размер MTU: %s, для созданного интерфейса: %s\n", ip, iface.Name())

	cmd := fmt.Sprintf("ip link set dev %s mtu %d", iface.Name(), mtu)
	out, err := command.RunCommand(cmd)
	if err != nil {
		log.Println(out)
		return nil, err
	}

	log.Printf("Назначаем IP адресс: %s, для созданного интерфейса: %s\n", ip, iface.Name())

	cmd = fmt.Sprintf("ip addr add %s/24 dev %s", ip, iface.Name())
	out, err = command.RunCommand(cmd)
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

	//log.Printf("Маршрутизируем пир: %s, для созданного интерфейса: %s\n", ip, iface.Name())
	//
	//cmd = fmt.Sprintf("ip addr add dev %s local %s peer %s", iface.Name(), ip, "10.1.1.2")
	//out, err = command.RunCommand(cmd)
	//if err != nil {
	//	log.Println(out)
	//	return nil, err
	//}
	//
	//log.Printf("Маршрутизируем подсеть через пир\n")
	//
	//cmd = fmt.Sprintf("ip route change %s via %s dev %s", "10.1.1.0/24", "10.1.1.2", iface.Name())
	//out, err = command.RunCommand(cmd)
	//if err != nil {
	//	log.Println(out)
	//	return nil, err
	//}

	return iface, nil
}

// ip link add ep1 type veth peer name ep221567890
