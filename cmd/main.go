package main

import (
	"log"
	"net"
	"netcasthub/config"
	"netcasthub/pkg"
	"time"

	"github.com/miekg/dns"
)

func main() {
	go pkg.Redirect()

	log.Printf("Trying to listen on the specific interface: %s\n", config.InterfaceName)

	iface, err := net.InterfaceByName(config.InterfaceName)
	if err != nil {
		log.Fatalf("Error finding interface '%s': %v. Make sure the name is correct and the interface is active.", config.InterfaceName, err)
	}

	// Set the address and port for mDNS.
	addr := &net.UDPAddr{
		IP:   net.ParseIP("224.0.0.251"),
		Port: 5353,
	}

	l, err := net.ListenMulticastUDP("udp4", iface, addr)
	if err != nil {
		log.Fatal("Make sure the interface is active and the multicast address is correct.")
	}

	handler := pkg.New(l)

	server := &dns.Server{
		PacketConn:  l,
		Handler:     dns.HandlerFunc(handler.HandleMDNSQuery),
		UDPSize:     4096,
		ReusePort:   true,
		ReuseAddr:   true,
		ReadTimeout: 5 * time.Second,
	}

	// Info to console
	log.Printf("mDNS server listening on 224.0.0.251:5353, bound to interface %s (%s)", iface.Name, iface.HardwareAddr)
	log.Println("Waiting for queries for '_googlecast._tcp.local.' and its subtypes...")

	err = server.ActivateAndServe()
	if err != nil {
		log.Fatalf("Error activating server: %v", err)
	}
	defer server.Shutdown()
}
