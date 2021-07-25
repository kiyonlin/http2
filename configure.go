package http2

import (
	"crypto/tls"
	"errors"
	"net"

	"github.com/valyala/fasthttp"
)

var (
	// ErrServerSupport indicates whether the server supports HTTP/2 or not.
	ErrServerSupport = errors.New("server doesn't support HTTP/2")
)

func configureDialer(d *Dialer) {
	if d.TLSConfig == nil {
		d.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
		}
	}

	tlsConfig := d.TLSConfig

	emptyServerName := len(tlsConfig.ServerName) == 0
	if emptyServerName {
		host, _, err := net.SplitHostPort(d.Addr)
		if err != nil {
			host = d.Addr
		}

		tlsConfig.ServerName = host
	}

	tlsConfig.NextProtos = append(tlsConfig.NextProtos, "h2")
}

// ConfigureClient configures the fasthttp.HostClient to run over HTTP/2.
func ConfigureClient(c *fasthttp.HostClient) error {
	emptyServerName := c.TLSConfig != nil && len(c.TLSConfig.ServerName) == 0

	d := &Dialer{
		Addr:      c.Addr,
		TLSConfig: c.TLSConfig,
	}

	c2, err := d.Dial()
	if err != nil {
		if err == ErrServerSupport && c.TLSConfig != nil { // remove added config settings
			for i := range c.TLSConfig.NextProtos {
				if c.TLSConfig.NextProtos[i] == "h2" {
					c.TLSConfig.NextProtos = append(c.TLSConfig.NextProtos[:i], c.TLSConfig.NextProtos[i+1:]...)
				}
			}

			if emptyServerName {
				c.TLSConfig.ServerName = ""
			}
		}

		return err
	}
	defer c2.Close()

	c.TLSConfig = d.TLSConfig

	cl := createClient(d)
	cl.conns.Init()

	c.Transport = cl.Do

	return nil
}

var ErrNotAvailableStreams = errors.New("ran out of available streams")