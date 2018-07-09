// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package session

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"time"
)

//SessionConfig contains settings for the session
type SessionService struct {
	Path           string
	HTTPOnly       bool
	Name           string
	Secure         bool
	IdleSessionTTL int64

	SessionProvider SessionProvider
}

func (sc SessionService) Create(rw http.ResponseWriter, r *http.Request) *Session {
	sid := base64.StdEncoding.EncodeToString(randomSecureKey(64))

	s := sc.SessionProvider.Create(sid)

	cookie := &http.Cookie{
		Path:     sc.Path,
		HttpOnly: sc.HTTPOnly,
		Name:     sc.Name,
		Secure:   sc.Secure,
		Value:    s.SessionID(),
	}

	http.SetCookie(rw, cookie)

	return s
}

//Get receives the session from the cookie
func (sc SessionService) Get(rw http.ResponseWriter, r *http.Request) (*Session, error) {
	cookie, err := r.Cookie(sc.Name)

	if err != nil {
		return nil, err
	}

	sess, err := sc.SessionProvider.Get(cookie.Value)

	if err != nil {
		//Try to remove client cookie as it is not valid anymore
		dc := &http.Cookie{
			Name:    sc.Name,
			MaxAge:  -1,
			Expires: time.Unix(1, 0),
			Path:    sc.Path,
		}

		http.SetCookie(rw, dc)

		return nil, err
	}

	return sess, nil
}

//Remove removes the session from the session map and the cookie
func (sc SessionService) Remove(rw http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(sc.Name)

	if err != nil {
		return err
	}

	sc.SessionProvider.Remove(cookie.Value)

	dc := &http.Cookie{
		Name:    sc.Name,
		MaxAge:  -1,
		Expires: time.Unix(1, 0),
		Path:    sc.Path,
	}

	http.SetCookie(rw, dc)

	return nil
}

//InitGC initialized the garbage collection for removing the session after the TTL has reached
func (sc SessionService) InitGC(ticker *time.Ticker, timeoutAfter time.Duration) {
	sc.SessionProvider.Clean(ticker, timeoutAfter)
}

func randomSecureKey(length int) []byte {
	k := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil
	}
	return k
}
