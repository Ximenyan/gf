// Copyright 2017 gf Author(https://github.com/gf). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gf.

// Package gredis provides convenient client for redis server.
//
// Redis Client.
//
// Redis Commands Official: https://redis.io/commands
//
// Redis Chinese Documentation: http://redisdoc.com/
package gredis

import (
	"fmt"
	"time"

	"github.com/gf/g/container/gmap"
	"github.com/gf/g/container/gvar"
	"github.com/gomodule/redigo/redis"
)

const (
	gDEFAULT_POOL_IDLE_TIMEOUT  = 60 * time.Second
	gDEFAULT_POOL_MAX_LIFE_TIME = 60 * time.Second
)

// Redis client.
type Redis struct {
	pool   *redis.Pool // Underlying connection pool.
	group  string      // Configuration group.
	config Config      // Configuration.
}

// Redis connection.
type Conn struct {
	redis.Conn
}

// Redis configuration.
type Config struct {
	Host            string
	Port            int
	Db              int
	Pass            string        // Password for AUTH.
	MaxIdle         int           // Maximum number of connections allowed to be idle (default is 0 means no idle connection)
	MaxActive       int           // Maximum number of connections limit (default is 0 means no limit)
	IdleTimeout     time.Duration // Maximum idle time for connection (default is 60 seconds, not allowed to be set to 0)
	MaxConnLifetime time.Duration // Maximum lifetime of the connection (default is 60 seconds, not allowed to be set to 0)
}

// Pool statistics.
type PoolStats struct {
	redis.PoolStats
}

var (
	// Instance map
	instances = gmap.NewStrAnyMap()
	// Pool map.
	pools = gmap.NewStrAnyMap()
)

// New creates a redis client object with given configuration.
// Redis client maintains a connection pool automatically.
func New(config Config) *Redis {
	if config.IdleTimeout == 0 {
		config.IdleTimeout = gDEFAULT_POOL_IDLE_TIMEOUT
	}
	if config.MaxConnLifetime == 0 {
		config.MaxConnLifetime = gDEFAULT_POOL_MAX_LIFE_TIME
	}
	return &Redis{
		config: config,
		pool: pools.GetOrSetFuncLock(fmt.Sprintf("%v", config), func() interface{} {
			return &redis.Pool{
				IdleTimeout:     config.IdleTimeout,
				MaxConnLifetime: config.MaxConnLifetime,
				Dial: func() (redis.Conn, error) {
					c, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
					if err != nil {
						return nil, err
					}
					// AUTH
					if len(config.Pass) > 0 {
						if _, err := c.Do("AUTH", config.Pass); err != nil {
							return nil, err
						}
					}
					// DB
					if _, err := c.Do("SELECT", config.Db); err != nil {
						return nil, err
					}
					return c, nil
				},
				// After the conn is taken from the connection pool, to test if the connection is available,
				// If error is returned then it closes the connection object and recreate a new connection.
				TestOnBorrow: func(c redis.Conn, t time.Time) error {
					_, err := c.Do("PING")
					return err
				},
			}
		}).(*redis.Pool),
	}
}

// Instance returns an instance of redis client with specified group.
// The <group> param is unnecessary, if <group> is not passed,
// it returns a redis instance with default group.
func Instance(name ...string) *Redis {
	group := DEFAULT_GROUP_NAME
	if len(name) > 0 {
		group = name[0]
	}
	v := instances.GetOrSetFuncLock(group, func() interface{} {
		if config, ok := GetConfig(group); ok {
			r := New(config)
			r.group = group
			return r
		}
		return nil
	})
	if v != nil {
		return v.(*Redis)
	}
	return nil
}

// Close closes the redis connection pool,
// it will release all connections reserved by this pool.
// It is not necessary to call Close manually.
func (r *Redis) Close() error {
	if r.group != "" {
		// If it is an instance object, it needs to remove it from the instance Map.
		instances.Remove(r.group)
	}
	pools.Remove(fmt.Sprintf("%v", r.config))
	return r.pool.Close()
}

// Conn returns a raw underlying connection object,
// which expose more methods to communicate with server.
// **You should call Close function manually if you do not use this connection any further.**
func (r *Redis) Conn() *Conn {
	return &Conn{r.pool.Get()}
}

// Alias of Conn, see Conn.
func (r *Redis) GetConn() *Conn {
	return r.Conn()
}

// SetMaxIdle sets the MaxIdle attribute of the connection pool.
func (r *Redis) SetMaxIdle(value int) {
	r.pool.MaxIdle = value
}

// SetMaxActive sets the MaxActive attribute of the connection pool.
func (r *Redis) SetMaxActive(value int) {
	r.pool.MaxActive = value
}

// SetIdleTimeout sets the IdleTimeout attribute of the connection pool.
func (r *Redis) SetIdleTimeout(value time.Duration) {
	r.pool.IdleTimeout = value
}

// SetMaxConnLifetime sets the MaxConnLifetime attribute of the connection pool.
func (r *Redis) SetMaxConnLifetime(value time.Duration) {
	r.pool.MaxConnLifetime = value
}

// Stats returns pool's statistics.
func (r *Redis) Stats() *PoolStats {
	return &PoolStats{r.pool.Stats()}
}

// Do sends a command to the server and returns the received reply.
// Do automatically get a connection from pool, and close it when reply received.
// It does not really "close" the connection, but drop it back to the connection pool.
func (r *Redis) Do(command string, args ...interface{}) (interface{}, error) {
	conn := &Conn{r.pool.Get()}
	defer conn.Close()
	return conn.Do(command, args...)
}

// DoVar returns value from Do as gvar.Var.
func (r *Redis) DoVar(command string, args ...interface{}) (*gvar.Var, error) {
	v, err := r.Do(command, args...)
	return gvar.New(v, true), err
}

// Deprecated.
// Send writes the command to the client's output buffer.
func (r *Redis) Send(command string, args ...interface{}) error {
	conn := &Conn{r.pool.Get()}
	defer conn.Close()
	return conn.Send(command, args...)
}
