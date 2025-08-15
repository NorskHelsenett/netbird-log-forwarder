package netbird

import (
	"fmt"
	"sync"

	"github.com/NorskHelsenett/netbird-log-forwarder/internal/logger"
	"github.com/NorskHelsenett/netbird-log-forwarder/pkg/models/netbird"
	"github.com/go-resty/resty/v2"
)

var GlobalUserCache *UserCache

type UserCache struct {
	mu        sync.RWMutex
	usersByID map[string]netbird.NetbirdUser
	token     string
	client    *resty.Client
}

func NewUserCache(token string) error {
	uc := &UserCache{
		usersByID: make(map[string]netbird.NetbirdUser),
		token:     token,
		client:    resty.New(),
	}
	if err := uc.refresh(); err != nil {
		return err
	}
	// return uc, nil
	GlobalUserCache = uc
	return nil
}

func (uc *UserCache) refresh() error {
	resp, err := uc.client.R().
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", "Token "+uc.token).
		SetResult(&[]netbird.NetbirdUser{}).
		Get("https://api.netbird.io/api/users")

	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("error response: %s", resp.Status())
	}

	users := *resp.Result().(*[]netbird.NetbirdUser)

	cache := make(map[string]netbird.NetbirdUser)
	for _, user := range users {
		cache[user.ID] = user
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.usersByID = cache

	logger.Log.Infoln("User cache refreshed")
	return nil
}

func (uc *UserCache) GetUserByID(id string) (netbird.NetbirdUser, error) {
	uc.mu.RLock()
	user, ok := uc.usersByID[id]
	uc.mu.RUnlock()

	if ok {
		return user, nil
	}

	if err := uc.refresh(); err != nil {
		return netbird.NetbirdUser{}, fmt.Errorf("refresh failed: %w", err)
	}

	uc.mu.RLock()
	defer uc.mu.RUnlock()
	user, ok = uc.usersByID[id]
	if !ok {
		return netbird.NetbirdUser{}, fmt.Errorf("user %q not found after refresh", id)
	}
	return user, nil
}
