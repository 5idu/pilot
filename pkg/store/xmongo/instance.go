package xmongo

import (
	"sync"
)

var _instances = sync.Map{}

// Range 遍历所有实例
func Range(fn func(name string, db *Client) bool) {
	_instances.Range(func(key, val interface{}) bool {
		return fn(key.(string), val.(*Client))
	})
}

// Get ...
func Get(name string) *Client {
	if ins, ok := _instances.Load(name); ok {
		return ins.(*Client)
	}
	return nil
}
