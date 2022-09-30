package fredis

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-redis/redis/v8"
	"oldjon.com/base/env"
)

type Mode uint8

const (
	Single Mode = iota
	Cluster
)

var (
	// ErrConfigNotFound is returned when redis config can not be found
	ErrConfigNotFound = errors.New("can not find redis config")

	// ErrTTLKeyNotExpireSet is returned  when ttl key exists but has no associated expire
	ErrTTLKeyNotExpireSet = errors.New("ttl key exists but has no associated expire")

	// ErrTTLKeyNotExist is returned when ttl key not exist
	ErrTTLKeyNotExist = errors.New("ttl key not exist")

	ErrRedisConfigNotFound = errors.New("can not find redis config")
)

const (
	ttlKeyNotExpireSet = -1
	ttlKeyNotExists    = -2
)

// RedisClientOption redis client option
type RedisClientOption struct {
	Mode                Mode
	Addr                string
	Password            string
	Db                  int
	PoolSize            int
	Addrs               map[string]string
	ClusterAddrs        []string
	ClusterMaxRedirects int
	ClusterReadOnly     bool
	DialTimeout         time.Duration
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
}

// Client introduce all the redis method we need for redis client and also with context support
type RedisClient interface {
	io.Closer
	Generic
	String
}

// NewRedisClient create a redisClient object from config
// it will create client connect to single, ring or cluster based on the configuration
// currently we only support redisClient, may add more in the future (for example , v9Client?), so we return an interface
func NewRedisClient(option *RedisClientOption) (RedisClient, error) {
	// keep backward compatible with current config file
	var client *redisClient
	var err error

	switch option.Mode {
	case Single:
		client, err = newRedisClientSingle(
			option.Addr,
			option.Password,
			option.Db,
			option.PoolSize,
			option.DialTimeout,
			option.ReadTimeout,
			option.WriteTimeout,
		)
	case Cluster:
		client, err = newRedisClientCluster(
			option.ClusterAddrs,
			option.Password,
			option.PoolSize,
			option.ClusterMaxRedirects,
			option.ClusterReadOnly,
			option.DialTimeout,
			option.ReadTimeout,
			option.WriteTimeout,
		)
	default:
		return nil, ErrConfigNotFound
	}

	if err != nil {
		return nil, err
	}

	// run a ping test?
	if _, err = client.client.Ping(context.TODO()).Result(); err != nil {
		return nil, err
	}

	return client, nil
}

type redisClient struct {
	client redis.UniversalClient // client would be a universal client to support single or ring or cluster
	mode   string
}

func getRedisMode(configReader env.ModuleConfigReader, configKey string) Mode {
	mode := configReader.GetString(configKey + ".mode")
	if mode == "single" {
		return Single
	}
	if mode == "cluster" {
		return Cluster
	}
	return Single
}

// newRedisClientSingle create a RedisClient object using redis v8 client in single instance mode
func newRedisClientSingle(addr string, password string, db int, poolSize int, dialTimeOut time.Duration,
	readTimeOut time.Duration, writeTimeOut time.Duration) (*redisClient, error) {
	rc := redisClient{}
	rc.client = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     poolSize,
		DialTimeout:  dialTimeOut,
		ReadTimeout:  readTimeOut,
		WriteTimeout: writeTimeOut,
	})

	rc.mode = "single"

	return &rc, nil
}

// newRedisClientSingle create a RedisClient object using redis v8 client in cluster instance mode
func newRedisClientCluster(addrs []string, password string, poolSize int,
	maxRedirects int, readOnly bool, dialTimeOut time.Duration,
	readTimeOut time.Duration, writeTimeOut time.Duration) (*redisClient, error) {
	rc := redisClient{}
	rc.client = redis.NewClusterClient(&redis.ClusterOptions{
		NewClient: func(opt *redis.Options) *redis.Client {
			node := redis.NewClient(opt)
			return node
		},
		Addrs:        addrs,
		Password:     password,
		PoolSize:     poolSize,
		MaxRedirects: maxRedirects,
		ReadOnly:     readOnly,
		DialTimeout:  dialTimeOut,
		ReadTimeout:  readTimeOut,
		WriteTimeout: writeTimeOut,
	})

	rc.mode = "cluster"
	return &rc, nil
}

// Close redis connection
func (rc *redisClient) Close() error {
	return rc.client.Close()
}

func NewRedisClientByConfig(cfg env.ModuleConfigReader, redisDBKey string) (RedisClient, error) {
	var redisConfig RedisClientOption

	redisMode := getRedisMode(cfg, redisDBKey)
	switch redisMode {
	case Single:
		redisConfig = RedisClientOption{
			Mode: Single,
			Addr: cfg.GetString(redisDBKey + ".addr"),
			Db:   cfg.GetInt(redisDBKey + ".db"),
		}
	case Cluster:
		redisConfig = RedisClientOption{
			Mode:                Cluster,
			ClusterAddrs:        cfg.GetStringSlice(redisDBKey + ".addrs"),
			ClusterMaxRedirects: cfg.GetInt(redisDBKey + ".maxredirects"),
			ClusterReadOnly:     cfg.GetBool(redisDBKey + ".readonly"),
		}
	default:
		return nil, fmt.Errorf("%w: mode[%d]", ErrRedisConfigNotFound, redisMode)
	}

	// set common config
	redisConfig.PoolSize = cfg.GetInt(redisDBKey + ".poolsize")
	redisConfig.Password = cfg.GetString(redisDBKey + ".password")
	redisConfig.ReadTimeout = time.Duration(cfg.GetInt(redisDBKey+".readtimeout")) * time.Second

	client, err := NewRedisClient(&redisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w, %s aaaaaaa", err, redisConfig.Addr)
	}

	return client, nil
}
