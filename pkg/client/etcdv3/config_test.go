package etcdv3

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/5idu/pilot/pkg/conf"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestDefaultConfig(t *testing.T) {
	defaultConfig := DefaultConfig()
	assert.Equal(t, time.Second*5, defaultConfig.ConnectTimeout)
	assert.Equal(t, false, defaultConfig.BasicAuth)
	assert.Equal(t, []string{"http://localhost:2379"}, defaultConfig.Endpoints)
	assert.Equal(t, false, defaultConfig.Secure)
}

func TestConfigSet(t *testing.T) {
	config := DefaultConfig()
	config.Endpoints = []string{"localhost"}
	assert.Equal(t, []string{"localhost"}, config.Endpoints)
}

func TestDelKeys(t *testing.T) {
	err := conf.LoadFromReader(bytes.NewBufferString(""), yaml.Unmarshal)
	if err != nil {
		panic(err)
	}

	config := DefaultConfig()
	config.Endpoints = []string{"http://localhost:12379", "http://localhost:22379", "http://localhost:32379"}
	config.MustSingleton().DelKeys(context.Background(), []string{"n1", "n2", "n3"}...)
}
