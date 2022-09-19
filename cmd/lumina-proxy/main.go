package main

import (
	"crypto/tls"
	"flag"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/c0va23/go-proxyprotocol"
	"github.com/zhangyoufu/lumina"
)

func main() {
	var (
		enableTls      bool
		enableProxy    bool
		listenAddr     string
		tlsCertPath    string
		tlsKeyPath     string
		serverListPath string
		ln             net.Listener
	)
	flag.BoolVar(&enableTls, "tls", false, "enable TLS")
	flag.BoolVar(&enableProxy, "proxy", false, "enable PROXY protocol support")
	flag.StringVar(&listenAddr, "listen", ":8000", "listen address")
	flag.StringVar(&tlsCertPath, "tlsCert", "cert.pem", "path to TLS certificate (PEM format)")
	flag.StringVar(&tlsKeyPath, "tlsKey", "key.pem", "path to TLS certificate key (PEM format)")

	flag.StringVar(&serverListPath, "serverlist", "server.yaml", "path to list of server.yaml")
	flag.Parse()

	tcpListener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal("unable to listen: ", err)
	}
	ln = tcpListener

	if enableTls {
		cert, err := tls.LoadX509KeyPair(tlsCertPath, tlsKeyPath)
		if err != nil {
			log.Fatal("unable to load X509 key pair: ", err)
		}
		config := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		tlsListener := tls.NewListener(ln, config)
		ln = tlsListener
	}

	if enableProxy {
		// DefaultFallbackHeaderParserBuilder contains StubHeaderParserBuilder,
		// which accepts non-PROXY protocol traffic
		proxyListener := proxyprotocol.NewDefaultListener(ln)
		ln = proxyListener
	}

	f, err := os.Open(serverListPath)
	if err != nil {
		log.Fatal("unable to open server.yaml: ", err)
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal("unable to read server.yaml: ", err)
	}

	type ServerEntry struct {
		Name string
		Addr string
		Cert string

		Key string
	}
	servers := make([]ServerEntry, 0)
	err = yaml.Unmarshal(data, &servers)
	if err != nil {
		log.Fatal("unable to parse server.yaml: ", err)
	}
	clients := make([]*lumina.Client, 0)
	for _, serverEntry := range servers {
		licKey := lumina.LicenseKey(serverEntry.Key)
		idaInfo := licKey.GetIDAInfo()
		licId := idaInfo.Id
		var dialer lumina.Dialer
		if serverEntry.Cert == "" {
			dialer = &lumina.TCPDialer{
				Addr: serverEntry.Addr,
			}
		} else {
			dialer = lumina.NewTLSDialer(serverEntry.Addr, serverEntry.Cert)
		}
		clients = append(clients, &lumina.Client{
			LicenseKey: licKey,
			LicenseId:  licId,
			Dialer:     dialer,
		})
	}

	proxy := NewProxyEx(clients)

	log.Print("proxy is listening on ", ln.Addr())
	log.Fatal("proxy stopped serving with error: ", proxy.Serve(ln))
	// TODO: graceful shutdown
}
