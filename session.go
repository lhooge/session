// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package session
package session

import (
	"sync"
	"time"
)

// Session contains the session
type Session struct {
	sid         string
	lastTouched time.Time

	values map[string]interface{}
	mutex  sync.RWMutex
}

// GetLastTouchTime receives the date when the session was touched last time
func (s *Session) GetLastTouchTime() time.Time {
	return s.lastTouched
}

// SessionID gets the sessionID
func (s *Session) SessionID() string {
	return s.sid
}

// SetValue sets a value into the session
func (s *Session) SetValue(key string, value interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.values[key] = value
}

// GetValue receives a value from the session
func (s *Session) GetValue(key string) interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.values[key]
}

// RemoveKey removes a key from the session
func (s *Session) RemoveKey(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.values, key)
}
