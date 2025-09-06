package server

import (
	"context"
	"sync"
)

type AnalysisEvent struct {
	AnalysisResponse
}

type Client struct {
	ConnId string
	UserId string
	Events chan AnalysisEvent
	ctx    context.Context
	cancel context.CancelFunc
}

type ConnManager struct {
	mu      sync.RWMutex
	clients map[string]map[string]*Client // map[userId]map[connId]Client
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		clients: make(map[string]map[string]*Client),
	}
}

func (cm *ConnManager) Get(userId, connId string) *Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if clients, exists := cm.clients[userId]; exists {
		if client, ok := clients[connId]; ok {
			return client
		}
	}
	return nil
}

func (cm *ConnManager) Subscribe(userId, connId string) *Client {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// User may already exist (multiple tabs open)
	if _, exists := cm.clients[userId]; !exists {
		cm.clients[userId] = make(map[string]*Client)
	}

	ctx, cancel := context.WithCancel(context.Background())
	client := &Client{
		ConnId: connId,
		UserId: userId,
		Events: make(chan AnalysisEvent, DefaultConfig.Engine.MaxDepth+1),
		ctx:    ctx,
		cancel: cancel,
	}
	cm.clients[userId][connId] = client
	return client
}

func (cm *ConnManager) Unsubscribe(userId, connId string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if clients, exists := cm.clients[userId]; exists {

		// Close and remove this connection
		if conn, ok := clients[connId]; ok && conn != nil {
			delete(clients, connId)
			close(conn.Events)
			conn.cancel()
		}

		// This user has no more connections, remove the user entry
		if len(clients) == 0 {
			delete(cm.clients, userId)
		}
	}
}

func (cm *ConnManager) Publish(userId, connId string, event AnalysisEvent) {
	c := cm.Get(userId, connId)
	if c != nil {
		select {
		case c.Events <- event:
		default:
			// Drop the event if the channel is full
		}
	}
}
