package proxy

import (
	"context"
	"shiroxy/pkg/models"

	"github.com/go-redis/redis/v8"
)

type ProxyServer struct {
	host   string
	port   string
	secure SecureMethod
}

type SecureMethod struct {
	IsMultipleTarget             bool
	CA_Certificate               string
	CA_KEY                       string
	Multiple_CertAndKey_Location string
	Redis_Host                   string
	Redis_Port                   string
	Redis_Password               string
	Redis_Connection_String      string
	RedisClient                  *redis.Client
}

func StartProxyServer(config *models.Config) (*ProxyServer, error) {
	var proxyServer ProxyServer
	// TODO: Implement Proxy Server Start Logic
	for i := 0; i < len(config.Frontend); i++ {
		frontend := config.Frontend[i]
		if frontend.Bind.Secure.Target == "single" {

		}
	}
	client, err := proxyServer.connectRedis()
	if err != nil {
		panic(err)
	}

	proxyServer.secure.RedisClient = client
	return &proxyServer, nil
}

func (p *ProxyServer) connectRedis() (*redis.Client, error) {

	var rdb redis.Client

	if p.secure.Redis_Connection_String != "" {
		opt, err := redis.ParseURL(p.secure.Redis_Connection_String)
		if err != nil {
			panic(err)
		}
		rdb = *redis.NewClient(opt)
	} else {
		rdb = *redis.NewClient(&redis.Options{
			Addr:     p.secure.Redis_Host + ":" + p.secure.Redis_Port,
			Password: p.secure.Redis_Password,
			DB:       0,
		})
	}

	var ctx context.Context = context.Background()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	} else {
		return &rdb, nil
	}
}
