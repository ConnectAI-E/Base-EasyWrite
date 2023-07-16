package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
)

type SupplierService struct {
	cache *cache.Cache
}
type SupplierCacheInterface interface {
	GetSuppliers() map[string]string
	SetSuppliers(data map[string]string)
	Clear() bool
}

var supplierService *SupplierService
var supplierCacheID = uuid.New().String()

func (p SupplierService) GetSuppliers() map[string]string {
	data, ok := p.cache.Get(productCacheID)
	if !ok {
		return nil
	}
	return data.(map[string]string)
}
func (p SupplierService) SetSuppliers(data map[string]string) {
	p.cache.Set(productCacheID, data, time.Hour*48)
}

func (p SupplierService) Clear() bool {
	p.cache.Delete(supplierCacheID)
	return true
}

func GetSupplierCache() SupplierCacheInterface {
	if supplierService == nil {
		supplierService = &SupplierService{cache: cache.New(48*time.Hour, 48*time.Hour)}
	}
	return supplierService
}
