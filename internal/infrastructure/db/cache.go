package database

import (
	"github.com/redis/go-redis/v9"
)

// CacheDatabase defines the contract for interacting with the cache storage system.
// It acts as a wrapper around the underlying Redis client to manage connections.
type CacheDatabase interface {
	// GetClient returns the underlying Redis client instance.
	// Use this when you need direct access to the Redis client to perform
	// specific commands or advanced operations.
	GetClient() *redis.Client

	// Close terminates the connection to the cache server.
	// It ensures that resources are released properly and returns an error
	// if the closure fails.
	Close() error
}
