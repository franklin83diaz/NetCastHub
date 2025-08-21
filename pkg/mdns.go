package pkg

import (
	"log"
	"net"
	"netcasthub/config"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

type DnsHandler struct {
	conn *net.UDPConn
	ip   string
}

func New(conn *net.UDPConn) *DnsHandler {
	return &DnsHandler{
		conn: conn,
	}
}

func (h *DnsHandler) HandleMDNSQuery(w dns.ResponseWriter, r *dns.Msg) {
	if config.Debug {
		log.Printf("--- Query mDNS from %s ---\n", w.RemoteAddr())
		log.Printf("  ID: %d, Recursive: %t, Authoritative: %t\n", r.Id, r.RecursionDesired, r.Authoritative)
		for _, q := range r.Question {
			log.Printf("  Question: Name: %s, Type: %s, Class: %s\n", q.Name, dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass])
		}
		log.Println("------------------------------------")
	}

	for _, q := range r.Question {
		//_%9E5E7C8F47989526C9BCD95D24084F6F0B27C5ED._sub._googlecast._tcp.local: type PTR, class IN, "QM" question
		if strings.HasSuffix(strings.ToLower(q.Name), "_googlecast._tcp.local") && q.Qtype == dns.TypePTR {
			if config.Debug {
				log.Printf(">>> Compatible query detected ('%s'). Preparing response...\n", q.Name)

			}

			m := new(dns.Msg)
			m.SetReply(r)
			m.Authoritative = true
			m.RecursionAvailable = false
			serviceInstance := config.MdnsCastDevices["serviceInstance"]
			serviceTarget := config.MdnsCastDevices["serviceTarget"]
			servicePort, _ := strconv.Atoi(config.MdnsCastDevices["servicePort"])
			serviceIP := config.MdnsCastDevices["serviceIP"]
			ptr := &dns.PTR{
				Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: 28800},
				Ptr: serviceInstance,
			}
			m.Answer = append(m.Answer, ptr)
			srv := &dns.SRV{
				Hdr:      dns.RR_Header{Name: serviceInstance, Rrtype: dns.TypeSRV, Class: dns.ClassINET, Ttl: 120},
				Priority: 0, Weight: 0, Port: uint16(servicePort), Target: serviceTarget,
			}
			txt := &dns.TXT{
				Hdr: dns.RR_Header{Name: serviceInstance, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 4500},
				Txt: []string{
					config.MdnsCastDevices["id"],
					config.MdnsCastDevices["cd"],
					config.MdnsCastDevices["rm"],
					config.MdnsCastDevices["ve"],
					config.MdnsCastDevices["md"],
					config.MdnsCastDevices["ic"],
					config.MdnsCastDevices["fn"],
					config.MdnsCastDevices["ca"],
					config.MdnsCastDevices["st"],
					config.MdnsCastDevices["bs"],
					config.MdnsCastDevices["nf"],
					config.MdnsCastDevices["rs"],
				},
			}
			a := &dns.A{
				Hdr: dns.RR_Header{Name: serviceTarget, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 120},
				A:   net.ParseIP(serviceIP),
			}
			nsec := &dns.NSEC{
				Hdr:        dns.RR_Header{Name: serviceInstance, Rrtype: dns.TypeNSEC, Class: dns.ClassINET, Ttl: 120},
				NextDomain: serviceInstance,
				TypeBitMap: []uint16{dns.TypeTXT, dns.TypeSRV},
			}
			m.Extra = append(m.Extra, srv, txt, a, nsec)

			// if err := w.WriteMsg(m); err != nil {
			//  log.Printf("Error writing response mDNS: %v\n", err)
			// } else {
			//  log.Println("<<< response for Google Cast sent successfully.")
			// }
			buf, err := m.Pack()
			if err != nil {
				log.Printf("Error packing DNS response: %v\n", err)
				return
			}
			_, err = h.conn.WriteTo(buf, &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353})
			if err != nil {
				log.Printf("Error writing mDNS response manually: %v\n", err)
			} else {
				log.Println("<<< response for Google Cast sent successfully (manually).")
			}
			return
		}
	}
}
