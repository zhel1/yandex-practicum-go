//
package server

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"github.com/zhel1/yandex-practicum-go/internal/certificates"
	"github.com/zhel1/yandex-practicum-go/internal/config"
	"net/http"
)

//Server struct
type Server struct {
	httpServer  *http.Server
	enableHTTPS bool
}

func NewServer(cfg *config.Config, handler http.Handler) *Server {
	server := &Server{
		enableHTTPS: cfg.EnableHTTPS,
		httpServer: &http.Server{
			Addr:    cfg.Addr,
			Handler: handler,
		},
	}

	if cfg.EnableHTTPS {
		cert, key := CreateServerCert()
		server.httpServer.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{{
				Certificate: [][]byte{cert.Raw},
				PrivateKey:  key,
				Leaf:        cert,
			}},
		}
	}

	return server
}

func (s *Server) Run() error {
	if s.enableHTTPS {
		return s.httpServer.ListenAndServeTLS("", "")
	} else {
		return s.httpServer.ListenAndServe()
	}
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func CreateServerCert() (*x509.Certificate, *rsa.PrivateKey) {
	rootCert, _, rootPriv := certificates.GenCARoot()
	DCACert, _, DCAPriv := certificates.GenDCA(rootCert, rootPriv)
	serverCert, _, ServerPriv := certificates.GenServerCert(DCACert, DCAPriv)
	return serverCert, ServerPriv
}
