package runtime

import "github.com/matthewmueller/joy/macro"

// Chan struct
type Chan struct {
	capacity int
	values   []interface{}
	sends    []interface{}
	recvs    []interface{}
	closed   bool
}

type send struct {
	value   interface{}
	promise interface{}
}

// Deferred fn
// TODO: cleanup
func Deferred() interface{} {
	macro.Raw(`
if (!(this instanceof Deferred)) return new Deferred()
var self = this

var p = new Promise(function(resolve, reject) {
	self.resolve = resolve
	self.reject = reject
})

self.then = p.then.bind(p)
self.catch = p.catch.bind(p)
`)
	return macro.Raw(`self`)
}

// Channel function
func Channel(capacity int) *Chan {
	return &Chan{
		capacity: capacity,
	}
}

// Send on the channel
func (c *Chan) Send(value interface{}) interface{} {
	if c.closed {
		return macro.Raw(`Promise.reject(new Error('send on a closed channel'))`)
	}

	// recv pending
	if len(c.recvs) > 0 {
		recv := c.recvs[0]
		_ = recv
		c.recvs = c.recvs[1:]
		macro.Raw(`recv.resolve(value)`)
		return macro.Raw(`Promise.resolve()`)
	}

	// room in buffer
	if len(c.values) < c.capacity {
		c.values = append(c.values, value)
		return macro.Raw(`Promise.resolve()`)
	}

	// no recv pending, block
	// TODO: right now this is new send(...)
	// should be fine for this case, just an FYI
	promise := Deferred()
	c.sends = append(c.sends, send{
		value:   value,
		promise: promise,
	})
	return promise
}

// Recv a value
func (c *Chan) Recv() interface{} {
	// values in buffer
	if len(c.values) > 0 {
		value := c.values[0]
		_ = value
		c.values = c.values[1:]
		return macro.Raw(`Promise.resolve(value)`)
	}

	// unblock pending sends
	if len(c.sends) > 0 {
		send := c.sends[0]
		_ = send
		c.sends = c.sends[1:]

		if c.closed {
			macro.Raw(`send.promise.reject(new Error('send on closed channel'))`)
			return macro.Raw(`Promise.resolve()`)
		}

		macro.Raw(`send.promise.resolve()`)
		return macro.Raw(`Promise.resolve(send.value)`)
	}

	// closed
	if c.closed {
		return macro.Raw(`Promise.resolve()`)
	}

	// no values, block
	promise := Deferred()
	c.recvs = append(c.recvs, promise)
	return promise
}

// Mapper interface
type Mapper interface {
	Set(key string, value interface{}) Mapper
	JSON() map[string]interface{}
}

type mapper struct {
	obj map[string]interface{}
}

// Map creates a new map
func Map() Mapper {
	return &mapper{}
}

func (m *mapper) Set(key string, value interface{}) Mapper {
	m.obj[key] = value
	return m
}

func (m *mapper) JSON() map[string]interface{} {
	return m.obj
}
