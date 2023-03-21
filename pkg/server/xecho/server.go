package xecho

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"reflect"

	"github.com/5idu/pilot/pkg/server"
	"github.com/5idu/pilot/pkg/xlog"

	"github.com/go-playground/locales/zh"
	translator "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// Server ...
type Server struct {
	*echo.Echo
	config   *Config
	listener net.Listener
	// registerer registry.Registry
}

func newServer(config *Config) (*Server, error) {
	var (
		listener net.Listener
		err      error
	)

	if config.EnableTLS {
		var cert, key []byte
		cert, err = os.ReadFile(config.CertFile)
		if err != nil {
			return nil, errors.Wrap(err, "read cert failed")
		}

		key, err = os.ReadFile(config.PrivateFile)
		if err != nil {
			return nil, errors.Wrap(err, "read private failed")
		}

		tlsConfig := new(tls.Config)
		tlsConfig.Certificates = make([]tls.Certificate, 1)

		if tlsConfig.Certificates[0], err = tls.X509KeyPair(cert, key); err != nil {
			return nil, errors.Wrap(err, "X509KeyPair failed")
		}
		listener, err = tls.Listen("tcp", config.Address(), tlsConfig)
	} else {
		listener, err = net.Listen("tcp", config.Address())
	}
	if err != nil {
		return nil, errors.Wrapf(err, "create xecho server failed")
	}
	config.Port = listener.Addr().(*net.TCPAddr).Port

	e := echo.New()
	e.Validator = NewCustomValidator()

	return &Server{
		Echo:     e,
		config:   config,
		listener: listener,
	}, nil
}

type CustomValidator struct {
	validator *validator.Validate
	trans     translator.Translator
}

func NewCustomValidator() *CustomValidator {
	validator := validator.New()
	uni := translator.New(zh.New())
	trans, _ := uni.GetTranslator("zh")
	// 注册一个函数，获取 struct tag 里自定义的 label 作为字段名
	validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		label := fld.Tag.Get("label")
		if label == "" {
			return fld.Name
		}
		return label
	})
	// 注册翻译器
	err := zh_translations.RegisterDefaultTranslations(validator, trans)
	if err != nil {
		xlog.Panic("register default translations failed", xlog.FieldErr(err))
	}
	return &CustomValidator{validator: validator, trans: trans}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	err := cv.validator.Struct(i)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return errors.New(err.Translate(cv.trans))
		}
	}
	return nil
}

func (s *Server) Healthz() bool {
	return true
}

// Serve implements server.Server interface.
func (s *Server) Serve() error {
	s.Echo.Logger.SetOutput(os.Stdout)
	s.Echo.Debug = s.config.Debug
	s.Echo.HideBanner = true
	// s.Echo.StdLogger = zap.NewStdLog(nil)
	if s.Echo.Debug {
		// display echo api list
		for _, route := range s.Echo.Routes() {
			fmt.Printf("[ECHO] \x1b[34m%8s\x1b[0m %s\n", route.Method, route.Path)
		}
	}

	var err error

	if s.config.EnableTLS {
		s.Echo.TLSListener = s.listener
		err = s.Echo.StartTLS("", s.config.CertFile, s.config.PrivateFile)
	} else {
		s.Echo.Listener = s.listener
		err = s.Echo.Start("")
	}

	if err != http.ErrServerClosed {
		return err
	}
	s.config.logger.Info("close echo", xlog.FieldExtra(map[string]interface{}{"addr": s.config.Address()}))
	return nil
}

// Stop implements server.Server interface
// it will terminate echo server immediately
func (s *Server) Stop() error {
	return s.Echo.Close()
}

// GracefulStop implements server.Server interface
// it will stop echo server gracefully
func (s *Server) GracefulStop(ctx context.Context) error {
	return s.Echo.Shutdown(ctx)
}

// Info returns server info, used by governor and consumer balancer
func (s *Server) Info() *server.ServiceInfo {
	serviceAddr := s.listener.Addr().String()
	if s.config.ServiceAddress != "" {
		serviceAddr = s.config.ServiceAddress
	}

	hostname, err := os.Hostname()
	if err != nil {
		s.config.logger.Error("info: get hostname error")
		return nil
	}

	info := server.ApplyOptions(
		server.WithName("echo.server"),
		server.WithScheme("http"),
		server.WithAddress(serviceAddr),
		server.WithHostname(hostname),
	)
	return &info
}
