package conf_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/5idu/pilot/pkg/conf"

	"gopkg.in/yaml.v3"
)

const mockConfData = `
redis:
  addr: 127.0.0.1:6379
  db: 0
  password:

etcd:
  endpoints: 
  - 127.0.0.1:12379
  - 127.0.0.1:22379
  - 127.0.0.1:32379

`

func processErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestConf(t *testing.T) {
	err := conf.LoadFromReader(bytes.NewBufferString(mockConfData), yaml.Unmarshal)
	processErr(t, err)
	fmt.Println(conf.Get("etcd.endpoints"))
}
