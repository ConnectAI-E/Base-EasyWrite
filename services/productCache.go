package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
)

type ProductService struct {
	cache *cache.Cache
}
type ProductCacheInterface interface {
	GetProducts() map[string][2]string
	SetProducts(data map[string][2]string)
	Clear() bool
}

var productService *ProductService
var productCacheID = uuid.New().String()

func (p ProductService) GetProducts() map[string][2]string {
	data, ok := p.cache.Get(productCacheID)
	if !ok {
		return nil
	}
	return data.(map[string][2]string)
}
func (p ProductService) SetProducts(data map[string][2]string) {
	p.cache.Set(productCacheID, data, time.Hour*48)
}

func (p ProductService) Clear() bool {
	p.cache.Delete(productCacheID)
	return true
}

func GetProductCache() ProductCacheInterface {
	if productService == nil {
		productService = &ProductService{cache: cache.New(48*time.Hour, 48*time.Hour)}
	}
	return productService
}
