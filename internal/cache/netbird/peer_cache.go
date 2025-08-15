package netbird

import (
	"fmt"
	"sync"

	"github.com/NorskHelsenett/netbird-log-forwarder/internal/logger"
	"github.com/NorskHelsenett/netbird-log-forwarder/pkg/models/netbird"
	"github.com/go-resty/resty/v2"
)

var GlobalPeerCache *PeerCache

type PeerCache struct {
	mu        sync.RWMutex
	peersByID map[string]netbird.NetbirdPeer
	token     string
	client    *resty.Client
}

func NewPeerCache(token string) error {
	pc := &PeerCache{
		token:  token,
		client: resty.New(),
	}
	if err := pc.refresh(); err != nil {
		return err
	}
	// return uc, nil
	GlobalPeerCache = pc
	return nil
}

func (pc *PeerCache) refresh() error {
	resp, err := pc.client.R().
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", "Token "+pc.token).
		SetResult(&[]netbird.NetbirdPeer{}).
		Get("https://api.netbird.io/api/peers")

	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("error response: %s", resp.Status())
	}

	peers := *resp.Result().(*[]netbird.NetbirdPeer)

	cache := make(map[string]netbird.NetbirdPeer)
	for _, peer := range peers {
		cache[peer.ID] = peer
	}

	// for _, peer := range peers {
	// 	cache[peer.ID] = peer
	// }

	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.peersByID = cache

	logger.Log.Infoln("Peer cache refreshed")
	return nil
}

func (pc *PeerCache) GetPeerByID(id string) (netbird.NetbirdPeer, error) {
	pc.mu.RLock()
	peer, ok := pc.peersByID[id]
	pc.mu.RUnlock()

	if ok {
		return peer, nil
	}

	if err := pc.refresh(); err != nil {
		return netbird.NetbirdPeer{}, fmt.Errorf("refresh failed: %w", err)
	}

	pc.mu.RLock()
	defer pc.mu.RUnlock()
	peer, ok = pc.peersByID[id]
	if !ok {
		return netbird.NetbirdPeer{}, fmt.Errorf("peer %q not found after refresh", id)
	}
	return peer, nil
}
