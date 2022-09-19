// +build go1.13

package lumina

import (
	"crypto/tls"
	"crypto/x509"
	"net"
)

// TLSDialer embeds net.Dialer and tls.Config.
type TLSDialer struct {
	TCPDialer
	tls.Config
}

// If RootCAs is nil, the host's root CA set will be used.
func (d *TLSDialer) Dial() (net.Conn, error) {
	return tls.DialWithDialer(&d.Dialer, "tcp", d.Addr, &d.Config)
}

func (d *TLSDialer) Info() string {
	return "Lumina-TLS-" + d.Addr
}

func NewTLSDialer(serverAddr string, cert string) Dialer {
	roots := x509.NewCertPool()
	if ok := roots.AppendCertsFromPEM([]byte(cert)); !ok {
		panic("unable to parse Hex-Rays cert")
	}
	d := &TLSDialer{}
	d.Addr = serverAddr
	d.RootCAs = roots
	d.MinVersion = tls.VersionTLS13

	return d
}
