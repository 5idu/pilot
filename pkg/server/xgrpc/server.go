package xgrpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"

	"github.com/5idu/pilot/pkg/server"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Server ...
type Server struct {
	*grpc.Server
	listener net.Listener
	*Config
}

func newServer(config *Config) (*Server, error) {
	var streamInterceptors = append(
		[]grpc.StreamServerInterceptor{defaultStreamServerInterceptor(config.logger, config)},
		config.streamInterceptors...,
	)

	var unaryInterceptors = append(
		[]grpc.UnaryServerInterceptor{defaultUnaryServerInterceptor(config.logger, config)},
		config.unaryInterceptors...,
	)

	if config.EnableTLS {
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.PrivateFile)
		if err != nil {
			return nil, errors.Wrap(err, "tls.LoadX509KeyPair failed")
		}

		certPool := x509.NewCertPool()
		rootBuf, err := os.ReadFile(config.CaFile)
		if err != nil {
			return nil, errors.Wrap(err, "os.ReadFile failed")
		}
		if !certPool.AppendCertsFromPEM(rootBuf) {
			return nil, errors.New("certPool.AppendCertsFromPEM failed")
		}

		tlsConf := &tls.Config{
			ClientAuth:   tls.RequireAndVerifyClientCert,
			Certificates: []tls.Certificate{cert},
			ClientCAs:    certPool,
		}

		config.serverOptions = append(config.serverOptions,
			grpc.Creds(credentials.NewTLS(tlsConf)),
		)
	}

	config.serverOptions = append(config.serverOptions,
		grpc.StreamInterceptor(StreamInterceptorChain(streamInterceptors...)),
		grpc.UnaryInterceptor(UnaryInterceptorChain(unaryInterceptors...)),
	)

	newServer := grpc.NewServer(config.serverOptions...)
	listener, err := net.Listen(config.Network, config.Address())
	if err != nil {
		return nil, errors.Wrap(err, "net.Listen failed")
	}
	config.Port = listener.Addr().(*net.TCPAddr).Port

	return &Server{
		Server:   newServer,
		listener: listener,
		Config:   config,
	}, nil
}

func (s *Server) Healthz() bool {
	return true
}

// Server implements server.Server interface.
func (s *Server) Serve() error {
	// display grpc server method list
	for fm, info := range s.GetServiceInfo() {
		for _, method := range info.Methods {
			fmt.Printf("[GRPC] \x1b[34m%8s\x1b[0m.%s\n", fm, method.Name)
		}
	}
	// display grpc server addr
	fmt.Printf("[GRPC] \x1b[33m%8s\x1b[0m %s\n", "Listen On", s.listener.Addr().String())
	err := s.Server.Serve(s.listener)
	return err
}

// Stop implements server.Server interface
// it will terminate echo server immediately
func (s *Server) Stop() error {
	s.Server.Stop()
	return nil
}

// GracefulStop implements server.Server interface
// it will stop echo server gracefully
func (s *Server) GracefulStop(ctx context.Context) error {
	s.Server.GracefulStop()
	return nil
}

// Info returns server info, used by governor and consumer balancer
func (s *Server) Info() *server.ServiceInfo {
	serviceAddress := s.listener.Addr().String()
	if s.Config.ServiceAddress != "" {
		serviceAddress = s.Config.ServiceAddress
	}

	info := server.ApplyOptions(
		server.WithScheme("grpc"),
		server.WithAddress(serviceAddress),
	)
	return &info
}
