package lib

import (
	"sync"
)

// TokenBlacklist 用于管理已失效的token
type TokenBlacklist struct {
	blacklist map[string]bool
	mutex     sync.RWMutex
}

var (
	blacklist *TokenBlacklist
	once      sync.Once
)

// GetTokenBlacklist 获取token黑名单单例
func GetTokenBlacklist() *TokenBlacklist {
	once.Do(func() {
		blacklist = &TokenBlacklist{
			blacklist: make(map[string]bool),
		}
	})
	return blacklist
}

// AddToBlacklist 将token加入黑名单
func (t *TokenBlacklist) AddToBlacklist(token string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.blacklist[token] = true
}

// IsBlacklisted 检查token是否在黑名单中
func (t *TokenBlacklist) IsBlacklisted(token string) bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.blacklist[token]
} 