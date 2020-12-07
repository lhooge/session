// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package session

import (
	"fmt"
	"sync"
	"time"
)

//SessionProvider an interface for storing and accessing sessions
type SessionProvider interface {
	Create(sid string) *Session
	Get(sid string) (*Session, error)
	FindSessionsByValue(key string, value interface{}) []Session
	Remove(sid string)
	Clean(ticker *time.Ticker, timeoutAfter time.Duration)
}

//InMemoryProvider implements a in memory storage for sessions
type InMemoryProvider struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
}

//NewInMemoryProvider creates a new in memory provider
func NewInMemoryProvider() *InMemoryProvider {
	return &InMemoryProvider{
		sessions: make(map[string]*Session),
	}
}

//Create stores a session in the map
func (imp *InMemoryProvider) Create(sid string) *Session {
	imp.mutex.Lock()

	defer imp.mutex.Unlock()

	imp.sessions[sid] = &Session{
		sid:         sid,
		lastTouched: time.Now(),
		values:      make(map[string]interface{}),
	}

	return imp.sessions[sid]
}

//Get receives the session from the map by the session identifier
func (imp *InMemoryProvider) Get(sid string) (*Session, error) {
	imp.mutex.RLock()
	defer imp.mutex.RUnlock()
	if sess, ok := imp.sessions[sid]; ok {
		sess.lastTouched = time.Now()
		return sess, nil
	}

	return nil, fmt.Errorf("no session with id %s found", sid)
}

//FindSessionsByValues finds all sessions from the map found by the key and value
func (imp *InMemoryProvider) FindSessionsByValue(key string, value interface{}) []Session {
	imp.mutex.RLock()

	defer imp.mutex.RUnlock()

	var sessions []Session

	for _, s := range imp.sessions {
		if s.values[key] == value {
			sessions = append(sessions, *s)
		}
	}

	return sessions
}

//Remove removes a session by the session identifier from the map
func (imp *InMemoryProvider) Remove(sid string) {
	imp.mutex.Lock()
	defer imp.mutex.Unlock()

	delete(imp.sessions, sid)
}

//Clean clean sessions after the specified timeout
func (imp *InMemoryProvider) Clean(ticker *time.Ticker, timeoutAfter time.Duration) {
	go func() {
		for range ticker.C {
			imp.mutex.Lock()
			for key, value := range imp.sessions {
				if time.Now().After(value.lastTouched.Add(timeoutAfter)) {
					delete(imp.sessions, key)
				}
			}
			imp.mutex.Unlock()
		}
	}()
}
