// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package session_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"git.hoogi.eu/snafu/session"
)

func sampleSessionService() session.SessionService {
	sessionService := session.SessionService{
		Secure:         true,
		Path:           "/test",
		HTTPOnly:       true,
		Name:           "test-session",
		IdleSessionTTL: 1,

		SessionProvider: session.NewInMemoryProvider(),
	}

	return sessionService
}

func TestSessionLifeCycle(t *testing.T) {
	sessionService := sampleSessionService()

	dur, _ := time.ParseDuration("1s")
	ticker := time.NewTicker(dur)
	sessionService.InitGC(ticker, 1*1e9)

	session, cookie := createSession(t, sessionService)

	expectedCookie := &http.Cookie{
		Name:     sessionService.Name,
		HttpOnly: sessionService.HTTPOnly,
		Secure:   sessionService.Secure,
		Path:     "/test",

		Value: session.SessionID(),
	}

	session.SetValue("userid", 3)
	checkCookie(t, cookie, expectedCookie)

	getSess, err := getSession(t, cookie.Raw, sessionService)

	if err != nil {
		t.Fatal(err)
	}

	if getSess.SessionID() != session.SessionID() {
		t.Fatalf("invalid session id, want %s, got %s", session.SessionID(), getSess.SessionID())
	}

	if getSess.GetValue("userid") != 3 {
		t.Fatalf("the session does not contain the expected user id %d in the session values, got userid %v", 3, session.GetValue("userid"))
	}

	removeSession(t, cookie.Raw, sessionService)

	_, err = getSession(t, cookie.Raw, sessionService)

	if err == nil {
		t.Fatalf("the session should be removed, but is still there %v", err)
	}
}

func TestSessionGarbageCollection(t *testing.T) {
	sessionService := sampleSessionService()

	ticker := time.NewTicker(time.Duration(1 * 1e9))
	sessionService.InitGC(ticker, time.Duration(2*time.Second))

	session, cookie := createSession(t, sessionService)

	expectedCookie := &http.Cookie{
		Name:     sessionService.Name,
		HttpOnly: sessionService.HTTPOnly,
		Secure:   sessionService.Secure,
		Path:     "/test",

		Value: session.SessionID(),
	}

	checkCookie(t, cookie, expectedCookie)

	time.Sleep(time.Duration(3 * 1e9))

	getSess, err := getSession(t, cookie.Raw, sessionService)
	if err == nil {
		t.Fatal(err)
	}

	if getSess != nil {
		t.Fatalf("got a session which should be invalidated. Initial %s", session.SessionID())
	}
}

func createSession(t *testing.T, sc session.SessionService) (*session.Session, *http.Cookie) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rw := httptest.NewRecorder()

	createdSession := sc.Create(rw, req)
	cookies := rw.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookies received")
	}
	return createdSession, cookies[0]
}

func getSession(t *testing.T, rawCookieValue string, sc session.SessionService) (*session.Session, error) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Cookie", rawCookieValue)
	rw := httptest.NewRecorder()
	getSession, err := sc.Get(rw, req)
	return getSession, err
}

func removeSession(t *testing.T, rawCookieValue string, cs session.SessionService) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Cookie", rawCookieValue)
	rw := httptest.NewRecorder()
	if err := cs.Remove(rw, req); err != nil {
		t.Fatal(err)
	}
}

func checkCookie(t *testing.T, cookie, expectedCookie *http.Cookie) {
	if cookie.Name != expectedCookie.Name {
		t.Errorf("unexpected cookie name, want: %s, got: %s", expectedCookie.Name, cookie.Name)
	}
	if cookie.Path != expectedCookie.Path {
		t.Errorf("unexpected cookie path, want: %s, got %s", expectedCookie.Path, cookie.Path)
	}
	if cookie.Value != expectedCookie.Value {
		t.Errorf("unexpected cookie value, want %s, got %s", expectedCookie.Value, cookie.Value)
	}
	if cookie.HttpOnly != expectedCookie.HttpOnly {
		t.Errorf("unexpected cookie http only flag, want %t, got %t", expectedCookie.HttpOnly, cookie.HttpOnly)
	}
	if cookie.Secure != expectedCookie.Secure {
		t.Errorf("unexpected cookie secure flag, want %t, got %t", expectedCookie.Secure, cookie.Secure)
	}
}
