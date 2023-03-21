package xecho_test

import (
	"bytes"
	"context"

	_ "github.com/5idu/pilot/pkg/registry/etcdv3"

	"net/http"
	"testing"
	"time"

	"github.com/5idu/pilot/pkg/conf"
	"github.com/5idu/pilot/pkg/registry"
	"github.com/5idu/pilot/pkg/server/xecho"
	"github.com/5idu/pilot/pkg/xlog"

	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

const configData = `
pilot:
  server:
    http:
      port: 9091
  
  registry:
    default:
      endpoints:
        - 127.0.0.1:12379
        - 127.0.0.1:22379
        - 127.0.0.1:32379

`

func Test_Server(t *testing.T) {
	// new server
	err := conf.LoadFromReader(bytes.NewBufferString(configData), yaml.Unmarshal)
	if err != nil {
		t.Fatal(err)
	}

	c := xecho.StdConfig("http")
	c.Port = 0
	s := c.MustBuild()
	s.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"data": "hello pilot",
		})
	})
	s.GET("/api/github.com/5idu/pilot/biz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"data": "hello pilot",
		})
	})
	s.POST("/api/github.com/5idu/pilot/internal", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"data": "hello pilot",
		})
	})
	s.POST("/api/github.com/5idu/pilot/hello", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"data": "hello pilot",
		})
	})

	go func() {
		registry.DefaultRegisterer.RegisterService(context.Background(), s.Info())
		xlog.Info("start server")
		s.Serve()
	}()

	time.Sleep(5 * time.Minute)
	assert.True(t, s.Healthz())
	assert.NotNil(t, s.Info())
	registry.DefaultRegisterer.Close()
	s.Stop()
}
