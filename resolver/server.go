package resolver

import (
	"time"

	"github.com/miekg/dns"

	"github.com/ray-g/dnsproxy/logger"
	"github.com/ray-g/dnsproxy/stats"
)

// Server type
type Server struct {
	addr      string
	rTimeout  time.Duration
	wTimeout  time.Duration
	handler   *DNSHandler
	udpServer *dns.Server
	tcpServer *dns.Server
}

func NewServer(addr string, handler *DNSHandler) *Server {
	return &Server{
		addr:     addr,
		handler:  handler,
		rTimeout: 5 * time.Second,
		wTimeout: 5 * time.Second,
	}
}

// Run starts the server
func (s *Server) Run() {
	tcpHandler := dns.NewServeMux()
	tcpHandler.HandleFunc(".", s.handler.DoTCP)

	udpHandler := dns.NewServeMux()
	udpHandler.HandleFunc(".", s.handler.DoUDP)

	s.tcpServer = &dns.Server{Addr: s.addr,
		Net:          "tcp",
		Handler:      tcpHandler,
		ReadTimeout:  s.rTimeout,
		WriteTimeout: s.wTimeout}

	s.udpServer = &dns.Server{Addr: s.addr,
		Net:          "udp",
		Handler:      udpHandler,
		UDPSize:      65535,
		ReadTimeout:  s.rTimeout,
		WriteTimeout: s.wTimeout}

	go s.start(s.udpServer)
	go s.start(s.tcpServer)
}

func (s *Server) start(ds *dns.Server) {
	logger.Infof("start %s listener on %s", ds.Net, s.addr)

	if err := ds.ListenAndServe(); err != nil {
		logger.Fatalf("start %s listener on %s failed: %s", ds.Net, s.addr, err.Error())
	}
}

// Stop stops the server
func (s *Server) Stop() {
	stats.Deactivate()

	if s.udpServer != nil {
		s.udpServer.Shutdown()
	}
	if s.tcpServer != nil {
		s.tcpServer.Shutdown()
	}
}
