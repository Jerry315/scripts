package cert

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"go.uber.org/zap"
	"net"
	"time"
)

var SkipVerify = false

var UTC = false

var TimeoutSeconds = 3

type Cert struct {
	DomainName string   `json:"domainName"`
	IP         string   `json:"ip"`
	Issuer     string   `json:"issuer"`
	CommonName string   `json:"commonName"`
	SANs       []string `json:"sans"`
	NotBefore  string   `json:"notBefore"`
	NotAfer    string   `json:"notAfer"`
	Error      string   `json:"error"`
	certChain  []*x509.Certificate
}

var serverCert = func(hostPort, domainName string,log *zap.Logger) ([]*x509.Certificate, string, error) {

	d := &net.Dialer{
		Timeout: time.Duration(TimeoutSeconds) * time.Second,
	}
	conn, err := tls.DialWithDialer(d, "tcp", hostPort, &tls.Config{
		InsecureSkipVerify: SkipVerify,
		ServerName:         domainName,
	})
	if err != nil {
		log.Error(fmt.Sprintf("#%v",err))
		return []*x509.Certificate{&x509.Certificate{}}, "", err
	}
	defer conn.Close()

	addr := conn.RemoteAddr()
	ip, _, _ := net.SplitHostPort(addr.String())
	cert := conn.ConnectionState().PeerCertificates
	return cert, ip, nil
}

func NewCert(hostPort, domainName string,log *zap.Logger) *Cert {
	certChain, ip, err := serverCert(hostPort, domainName,log)
	if err != nil {
		log.Error(fmt.Sprintf("#%v",err))
		return &Cert{DomainName: domainName, Error: err.Error()}
	}
	cert := certChain[0]

	var loc *time.Location
	loc = time.Local
	if UTC {
		loc = time.UTC
	}
	return &Cert{
		DomainName: domainName,
		IP:         ip,
		Issuer:     cert.Issuer.CommonName,
		CommonName: cert.Subject.CommonName,
		SANs:       cert.DNSNames,
		NotBefore:  cert.NotBefore.In(loc).String(),
		NotAfer:    cert.NotAfter.In(loc).String(),
		Error:      "",
		certChain:  certChain,
	}
}
