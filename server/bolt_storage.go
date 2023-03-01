package server

// See: https://github.com/etcd-io/bbolt

import (
	"fmt"
	"log"
	"net/url"
	"time"

	bolt "go.etcd.io/bbolt"
)

type BoltStorage struct {
	db *bolt.DB
}

const proxiesBucketName = "proxies"

func NewBoltStorage(path string) (*BoltStorage, error) {
	log.Printf("Opening BoltDB file at: %s", path)
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(proxiesBucketName))
		return err
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create proxies bucket %w", err)
	}

	return &BoltStorage{db}, nil
}

func (b *BoltStorage) Close() {
	log.Print("Closing BoltDB file")
	b.db.Close()
}

func getProxiesBucket(tx *bolt.Tx) *bolt.Bucket {
	bucket := tx.Bucket([]byte(proxiesBucketName))
	if bucket == nil {
		log.Fatal("proxies bucket wasn't created")
	}

	return bucket
}

func (s *BoltStorage) GetProxyDetails(host string) (proxy ProxyUrl, found bool, err error) {
	dbErr := s.db.View(func(tx *bolt.Tx) error {
		bucket := getProxiesBucket(tx)

		proxyBytes := bucket.Get([]byte(host))
		var err error
		if proxyBytes != nil {
			proxy, err = url.ParseRequestURI(string(proxyBytes))
			found = true
		} else {
			proxy = nil
			found = false
		}

		return err
	})

	if dbErr != nil {
		err = fmt.Errorf("failed to get proxy details: %w", dbErr)
	}

	return
}

func (s *BoltStorage) GetAll() (proxies []ProxyUrl, err error) {
	dbErr := s.db.Update(func(tx *bolt.Tx) error {
		bucket := getProxiesBucket(tx)
		return bucket.ForEach(func(_, v []byte) error {
			proxy, err := url.ParseRequestURI(string(v))
			if err != nil {
				return err
			}

			proxies = append(proxies, proxy)
			return nil
		})
	})

	if dbErr != nil {
		err = fmt.Errorf("failed to set proxy: %w", dbErr)
	}

	return
}

func (s *BoltStorage) SetProxy(host string, proxy ProxyUrl) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		bucket := getProxiesBucket(tx)
		return bucket.Put([]byte(host), []byte(proxy.String()))
	})

	if err != nil {
		return fmt.Errorf("failed to set proxy: %w", err)
	}
	return nil
}
