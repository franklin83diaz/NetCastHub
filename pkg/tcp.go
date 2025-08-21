package pkg

import (
	"fmt"
	"io"
	"log"
	"net"
	"netcasthub/config"
	"time"
)

func pipe(a, b net.Conn) {
	defer a.Close()
	defer b.Close()
	_ = a.SetDeadline(time.Now().Add(5 * time.Minute))
	_ = b.SetDeadline(time.Now().Add(5 * time.Minute))
	go io.Copy(a, b)
	io.Copy(b, a)
}

func Redirect() {
	ln, err := net.Listen("tcp", "0.0.0.0:8009")
	if err != nil {
		log.Fatal(err)
	}
	ipDevice := config.MdnsCastDevices["serviceIP"]
	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		fmt.Println("New connection from: ", c.RemoteAddr())
		go func() {
			up, err := net.Dial("tcp", ipDevice+":8009")
			if err != nil {
				c.Close()
				return
			}
			pipe(c, up)
		}()
	}
}
