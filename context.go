package gohelix

import "sync"

type Context struct {
	data map[string]interface{}
	sync.RWMutex
}

func NewContext() *Context {
	c := Context{
		data: make(map[string]interface{}),
	}
	return &c
}

func (c *Context) Set(key string, value interface{}) {
	c.Lock()
	c.data[key] = value
	c.Unlock()
}

func (c *Context) Get(key string) interface{} {
	v, ok := c.data[key]
	if !ok {
		return nil
	}

	return v
}
