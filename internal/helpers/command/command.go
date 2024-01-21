package command

import (
	"bytes"
	"fmt"
	"golang.org/x/net/ipv4"
	"log"
	"os/exec"
)

func RunCommand(cmd string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	c := exec.Command("bash", "-c", cmd)
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()

	log.Printf("Runned command: %s", cmd)

	if err != nil {
		return stderr.String(), fmt.Errorf("command error: %v", err)
	}

	return stdout.String(), err
}

func WritePacket(frame []byte) {
	header, err := ipv4.ParseHeader(frame)
	if err != nil {
		fmt.Println("write packet err:", err)
	} else {
		fmt.Println("SRC:", header.Src)
		fmt.Println("DST:", header.Dst)
		fmt.Println("ID:", header.ID)
		fmt.Println("CHECKSUM:", header.Checksum)
	}
}
