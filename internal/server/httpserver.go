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

// HttpServer struct
type HTTPServer struct {
	httpServer  *http.Server
	enableHTTPS bool
}

func NewHTTPServer(cfg *config.Config, handler http.Handler) *HTTPServer {
	server := &HTTPServer{
		enableHTTPS: cfg.EnableHTTPS,
		httpServer: &http.Server{
			Addr:    cfg.Addr,
			Handler: handler,
		},
	}

	if cfg.EnableHTTPS {
		cert, key := createServerCert()
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

func (s *HTTPServer) Run() error {
	if s.enableHTTPS {
		return s.httpServer.ListenAndServeTLS("", "")
	} else {
		return s.httpServer.ListenAndServe()
	}
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func createServerCert() (*x509.Certificate, *rsa.PrivateKey) {
	rootCert, _, rootPriv := certificates.GenCARoot()
	DCACert, _, DCAPriv := certificates.GenDCA(rootCert, rootPriv)
	serverCert, _, ServerPriv := certificates.GenServerCert(DCACert, DCAPriv)
	return serverCert, ServerPriv
}
