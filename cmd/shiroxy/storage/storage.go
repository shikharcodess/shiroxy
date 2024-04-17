package storage

import (
	"context"
	"errors"
	"shiroxy/pkg/models"

	"github.com/go-redis/redis/v8"
	"google.golang.org/protobuf/proto"
)

type Storage struct {
	storage        *models.Storage
	redisClient    *redis.Client
	domainMetadata map[string]*DomainMetadata
}

func InitializeStorage(storage *models.Storage) (*Storage, error) {
	storageSystem := Storage{
		storage: storage,
	}
	if storage.Location == "redis" {
		redisClient, err := storageSystem.connectRedis()
		if err != nil {
			return nil, err
		}

		storageSystem.redisClient = redisClient
	} else if storage.Location == "memory" {
		memoryStorage, err := storageSystem.initiazeMemoryStorage()
		if err != nil {
			return nil, err
		}

		storageSystem.domainMetadata = memoryStorage
	}
	return nil, nil
}

func (s *Storage) RegisterDomain(domainName string, createBody *DomainMetadata) error {
	if len(domainName) == 0 {
		return errors.New("domainName should not be empty")
	}

	if s.storage.Location == "memory" {
		s.domainMetadata[domainName] = createBody
	} else if s.storage.Location == "redis" {
		marshaledBody, err := proto.Marshal(createBody)
		if err != nil {
			return err
		}

		ctx := context.Background()
		result := s.redisClient.Set(ctx, domainName, marshaledBody, 0)
		if result.Err() != nil {
			return result.Err()
		}
	}
	return nil
}

func (s *Storage) UpdateDomain(domainName string, updateBody *DomainMetadata) error {
	if len(domainName) == 0 {
		return errors.New("domainName should not be empty")
	}

	if s.storage.Location == "memory" {
		oldData := s.domainMetadata[domainName]
		if oldData == nil {
			return errors.New("no data found for domainName")
		} else {
			s.domainMetadata[domainName] = updateBody
		}
	} else if s.storage.Location == "redis" {
		ctx := context.Background()
		result := s.redisClient.Get(ctx, domainName)
		if result.Err() != nil {
			return result.Err()
		}

		oldData, err := result.Result()
		if err != nil {
			return err
		}

		if oldData == "" {
			return errors.New("no data found for domainName")
		}
		marshaledUpdateBody, err := proto.Marshal(updateBody)
		if err != nil {
			return err
		}
		updateResult := s.redisClient.Set(ctx, domainName, marshaledUpdateBody, 0)
		if updateResult.Err() != nil {
			return updateResult.Err()
		}
	}
	return nil
}

func (s *Storage) RemoveDomain(domainName string) error {
	if len(domainName) == 0 {
		return errors.New("domainName should not be empty")
	}

	if s.storage.Location == "memory" {
		oldData := s.domainMetadata[domainName]
		if oldData == nil {
			return errors.New("no data found for domainName")
		} else {
			delete(s.domainMetadata, domainName)
		}
	} else if s.storage.Location == "redis" {
		ctx := context.Background()
		result := s.redisClient.Get(ctx, domainName)
		if result.Err() != nil {
			return result.Err()
		}

		oldData, err := result.Result()
		if err != nil {
			return err
		}

		if oldData == "" {
			return errors.New("no data found for domainName")
		}
		deleteResult := s.redisClient.Del(ctx, domainName)
		if deleteResult.Err() != nil {
			return deleteResult.Err()
		}
	}
	return nil
}

func (s *Storage) connectRedis() (*redis.Client, error) {
	var rdb redis.Client

	if s.storage.RedisConnectionString != "" {
		opt, err := redis.ParseURL(s.storage.RedisConnectionString)
		if err != nil {
			panic(err)
		}
		rdb = *redis.NewClient(opt)
	} else {
		rdb = *redis.NewClient(&redis.Options{
			Addr:     s.storage.RedisHost + ":" + s.storage.RedisHost,
			Password: s.storage.RedisPassword,
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

func (s *Storage) initiazeMemoryStorage() (map[string]*DomainMetadata, error) {
	memoryMap := map[string]*DomainMetadata{}
	return memoryMap, nil
}
