package gdb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/oldjon/gutil/env"
	"github.com/oldjon/gutil/marshaller"
	"github.com/opentracing/opentracing-go"
)

type Mode uint8

const (
	Single Mode = iota
	Cluster
)

var (
	// ErrConfigNotFound is returned when gredis config can not be found
	ErrConfigNotFound = errors.New("can not find gredis config")

	// ErrTTLKeyNotExpireSet is returned  when ttl key exists but has no associated expire
	ErrTTLKeyNotExpireSet = errors.New("ttl key exists but has no associated expire")

	// ErrTTLKeyNotExist is returned when ttl key not exist
	ErrTTLKeyNotExist = errors.New("ttl key not exist")

	ErrRedisConfigNotFound = errors.New("can not find gredis config")
)

const (
	ttlKeyNotExpireSet = -1
	ttlKeyNotExists    = -2
)

// RedisClientOption gredis client option
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
	Marshaller          gmarshaller.Marshaller
	Hooks               []redis.Hook
}

// RedisClient introduce all the gredis method we need for gredis client and also with context support
type RedisClient interface {
	io.Closer
	Generic
	String
	Hash
	SortedSet
	ObjectDB
	Scripter
	Pipeline() Pipeliner
	TxPipeline() Pipeliner
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
			option.Hooks,
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
			option.Hooks,
		)
	default:
		return nil, ErrConfigNotFound
	}

	if err != nil {
		return nil, err
	}

	client.objMarshaller = option.Marshaller

	// run a ping test?
	if _, err = client.client.Ping(context.TODO()).Result(); err != nil {
		return nil, err
	}

	return client, nil
}

type redisClient struct {
	client redis.UniversalClient // client would be a universal client to support single or ring or cluster
	mode   string
	// object
	objMarshaller gmarshaller.Marshaller
	koMapping     map[string]string
}

func getRedisMode(configReader env.ModuleConfig) Mode {
	mode := configReader.GetString("mode")
	if mode == "single" {
		return Single
	}
	if mode == "cluster" {
		return Cluster
	}
	return Single
}

// newRedisClientSingle create a RedisClient object using gredis v8 client in single instance mode
func newRedisClientSingle(addr string, password string, db int, poolSize int, dialTimeOut time.Duration,
	readTimeOut time.Duration, writeTimeOut time.Duration, hooks []redis.Hook) (*redisClient, error) {
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
	for _, hook := range hooks {
		rc.client.AddHook(hook)
	}
	rc.mode = "single"

	return &rc, nil
}

// newRedisClientSingle create a RedisClient object using gredis v8 client in cluster instance mode
func newRedisClientCluster(addrs []string, password string, poolSize int,
	maxRedirects int, readOnly bool, dialTimeOut time.Duration,
	readTimeOut time.Duration, writeTimeOut time.Duration, hooks []redis.Hook) (*redisClient, error) {
	rc := redisClient{}
	rc.client = redis.NewClusterClient(&redis.ClusterOptions{
		NewClient: func(opt *redis.Options) *redis.Client {
			node := redis.NewClient(opt)
			for _, hook := range hooks {
				node.AddHook(hook)
			}
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

// Close gredis connection
func (rc *redisClient) Close() error {
	return rc.client.Close()
}

func NewRedisClientByConfig(cfg env.ModuleConfig, marshaller string, tracer opentracing.Tracer) (RedisClient, error) {
	var redisConfig *RedisClientOption

	redisMode := getRedisMode(cfg)
	switch redisMode {
	case Single:
		redisConfig = &RedisClientOption{
			Mode: Single,
			Addr: cfg.GetString("addr"),
			Db:   cfg.GetInt("db"),
		}
	case Cluster:
		redisConfig = &RedisClientOption{
			Mode:                Cluster,
			ClusterAddrs:        cfg.GetStringSlice("addrs"),
			ClusterMaxRedirects: cfg.GetInt("maxredirects"),
			ClusterReadOnly:     cfg.GetBool("readonly"),
		}
	default:
		return nil, fmt.Errorf("%w: mode[%d]", ErrRedisConfigNotFound, redisMode)
	}

	// set common config
	redisConfig.PoolSize = cfg.GetInt("pool_size")
	redisConfig.Password = cfg.GetString("password")
	redisConfig.ReadTimeout = time.Duration(cfg.GetInt("readtimeout")) * time.Second
	if cfg.GetString("db_marshaller") != "" {
		marshaller = cfg.GetString("db_marshaller")
	}
	if marshaller == gmarshaller.MarshallerTypeJSON {
		redisConfig.Marshaller = &gmarshaller.JsonMarshaller{}
	} else if marshaller == gmarshaller.MarshallerTypeProtoBuf {
		redisConfig.Marshaller = &gmarshaller.ProtoMarshaller{}
	} else if marshaller == gmarshaller.MarshallerTypeProtoBufComp {
		redisConfig.Marshaller = &gmarshaller.ProtoCompressMarshaller{}
	} else { // default marshal by json
		marshaller = gmarshaller.MarshallerTypeJSON
		redisConfig.Marshaller = &gmarshaller.JsonMarshaller{}
	}

	if tracer != nil {
		// 增加 tracer hook
		redisConfig.Hooks = append(redisConfig.Hooks, &TraceHook{
			Tracer:     tracer,
			Instance:   redisConfig.Addr,
			RedisMode:  cfg.GetString("mode"),
			Marshaller: marshaller,
		})
	}

	client, err := NewRedisClient(redisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create gredis client: %w, %s", err, redisConfig.Addr)
	}

	return client, nil
}
