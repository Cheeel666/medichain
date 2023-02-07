package utils

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
)

func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func GetMyIP() string {
	var MyIP string

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatalln(err)
	} else {
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		MyIP = localAddr.IP.String()
	}
	return MyIP
}
