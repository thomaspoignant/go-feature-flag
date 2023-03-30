package apikey

import (
	"sync"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/internal/signer"
)

type Storage interface {
	Create(u ffuser.User) (string, error)
	Read(key string) (ffuser.User, bool)
	ReadAll() map[string]struct{}
	Delete(key string) error
}

type storageImpl struct {
	d map[string]ffuser.User
	m sync.RWMutex
}

func NewStorage(adminAPIKeys []string) Storage {
	adminAPIMap := make(map[string]ffuser.User, len(adminAPIKeys))
	for _, k := range adminAPIKeys {
		adminAPIMap[k] = ffuser.NewUserBuilder(k).
			Anonymous(false).
			AddCustom("admin", true).
			AddCustom("adminapikey", true).
			Build()
	}
	return &storageImpl{
		d: adminAPIMap,
		m: sync.RWMutex{},
	}
}

func (c *storageImpl) Create(u ffuser.User) (string, error) {
	key := signer.Sign([]byte(u.GetKey()), nil)
	c.m.Lock()
	c.d[key] = u
	c.m.Unlock()
	return key, nil
}

func (c *storageImpl) Read(key string) (ffuser.User, bool) {
	c.m.RLock()
	u, ok := c.d[key]
	c.m.RUnlock()
	return u, ok
}

func (c *storageImpl) ReadAll() map[string]struct{} {
	res := make(map[string]struct{}, len(c.d))
	c.m.RLock()
	for k := range c.d {
		res[k] = struct{}{}
	}
	c.m.RUnlock()
	return res
}

func (c *storageImpl) Delete(key string) error {
	c.m.Lock()
	delete(c.d, key)
	c.m.Unlock()
	return nil
}
