package cache

import "sync"

type Cacher interface {
	SetNewData(interface{})
	GetData() interface{}
}

// CachedData реализует простой кеш без ограничений по времени жизни
type CachedData struct {
	sync.RWMutex
	data interface{}
}

func NewCachedData() *CachedData {
	return &CachedData{data: nil}
}

func (app *CachedData) SetNewData(newData interface{}) {
	app.Lock()
	defer app.Unlock()

	app.data = newData
}

func (app *CachedData) GetData() interface{} {
	app.RLock()
	defer app.RUnlock()

	return app.data
}
