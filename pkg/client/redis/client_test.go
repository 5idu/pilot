package redis

import (
	"bytes"
	"context"
	"testing"

	"github.com/5idu/pilot/pkg/conf"

	"gopkg.in/yaml.v3"
)

var testconf = `
pilot:
  redis:
    default:
      master:
        addr: "redis://127.0.0.1:6379"
      db: 0
      poolSize: 2
      minIdleConns: 5
`

func TestClient(t *testing.T) {
	err := conf.LoadFromReader(bytes.NewBufferString(testconf), yaml.Unmarshal)
	if err != nil {
		t.Fatal(err)
	}

	cfg := StdConfig("default")
	client := cfg.MustBuild()
	t.Log(client.CmdOnMaster().Get(context.Background(), "name").String())
	client.CmdOnMaster().Set(context.Background(), "name", "bob", 0)
	t.Log(client.CmdOnMaster().Get(context.Background(), "name").String())
}
