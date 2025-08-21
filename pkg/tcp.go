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

	errChan := make(chan error, 2)

	go func() {
		_, err := io.Copy(b, a)
		errChan <- err
	}()

	go func() {
		_, err := io.Copy(a, b)
		errChan <- err
	}()

	err := <-errChan
	if err != nil && err != io.EOF {
		log.Printf("Error durante la copia de datos: %v", err)
	}
	fmt.Println("\033[32mConnection closed.\033[0m") // color green

}

func Redirect() {
	ln, err := net.Listen("tcp", "0.0.0.0:8009")
	if err != nil {
		log.Fatalf("Error to start listener: %v", err)
	}
	defer ln.Close()

	ipDevice := config.IpDevice
	log.Println("Listening on port 8009")

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Printf("Error to accept connection: %v", err)
			// if the error is temporary we can continue
			if ne, ok := err.(net.Error); ok && !ne.Temporary() {
				log.Fatalf("Critical error on listener: %v", err)
			}
			continue
		}
		//color blue
		fmt.Println("\033[34mNew connection from: ", c.RemoteAddr(), "\033[0m")

		go func(clientConn net.Conn) {
			// with timeout 10 seconds for avoiding long blocking
			up, err := net.DialTimeout("tcp", ipDevice+":8009", 10*time.Second)
			if err != nil {
				log.Printf("Could not connect to destination %s: %v", ipDevice, err)
				clientConn.Close()
				return
			}
			pipe(clientConn, up)
		}(c)
	}
}
