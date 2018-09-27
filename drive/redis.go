package drive

import (
	"github.com/go-redis/redis"
	"time"
	"ffmpeg/doHttp/tools/config"
)

func GetConn() *redis.Client {
	//根据配置文件读取redis登陆信息

	var redisH string
	var redisP string
	var redisPwd string

	redisHost, err := config.Config("redisHost")

	if err != nil {
		redisH = "127.0.0.1"
	} else {
		redisH = redisHost.(string)
	}

	redisPort, err := config.Config("redisPort")

	if err != nil {
		redisP = "6379"
	} else {
		redisP = redisPort.(string)
	}

	redisPassword, err := config.Config("redisPassword")

	if err != nil {
		redisPwd = "127.0.0.1"
	} else {
		redisPwd = redisPassword.(string)
	}

	return redis.NewClient(&redis.Options{
		Addr:     redisH + ":" + redisP,
		Password: redisPwd, // no password set
		DB:       0,        // use default DB
	})

	//return RedisClient
}

func RedisSetKeyValue(key string, val interface{}, expiration time.Duration) error {

	coon := GetConn()

	err := coon.Set(key, val, expiration).Err()

	defer coon.Close()

	return err
}

func RedisGetKeyValue(key string) (string, error) {

	coon := GetConn()

	res, err := coon.Get(key).Result()

	defer coon.Close()

	return res, err
}

func RedisLPush(key string, values interface{}) error {

	coon := GetConn()

	err := coon.LPush(key, values).Err()

	defer coon.Close()

	return err

}

func RedisRPush(key string, values interface{}) error {

	coon := GetConn()

	err := coon.RPush(key, values).Err()

	defer coon.Close()

	return err
}

func RedisRPop(key string) (string, error) {

	coon := GetConn()

	res, err := coon.RPop(key).Result()

	defer coon.Close()

	return res, err
}
