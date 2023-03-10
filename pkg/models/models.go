package models

type CacheManager interface {
	HSet(string, string, interface{}) error
	HGet(string, string) (string, error)
}
