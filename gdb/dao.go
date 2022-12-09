package gdb

type DB struct {
	ObjectDB // support redis db or mongo db(TODO)
	RedisClient
}

func NewDAO(objDB ObjectDB, redisClient RedisClient) *DB {
	return &DB{
		ObjectDB:    objDB,
		RedisClient: redisClient,
	}
}
