// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package session_test

import (
	"git.hoogi.eu/snafu/session"
	"testing"
)

func TestInMemoryStore(t *testing.T) {
	prov := session.NewInMemoryProvider()

	s1 := prov.Create("1234567890")
	s1.SetValue("userid", 8)

	s2 := prov.Create("1234567891")
	s2.SetValue("userid", 8)

	s3 := prov.Create("1234567892")
	s3.SetValue("userid", 8)

	s4 := prov.Create("1234567893")
	s4.SetValue("userid", 7)

	sessions := prov.FindByValue("userid", 8)

	if len(sessions) != 3 {
		t.Errorf("invalid length of sessions expected '3' bot got '%d'", len(sessions))
	}

	prov.Remove("1234567892")

	sessions = prov.FindByValue("userid", 8)

	if len(sessions) != 2 {
		t.Errorf("invalid length of sessions expected '2' bot got '%d'", len(sessions))
	}

	sess, err := prov.Get("1234567893")

	if err != nil {
		t.Error(err)
	}

	if sess.SessionID() != "1234567893" {
		t.Errorf("invalid session id returned expected '1234567893' bot got '%s'", sess.SessionID())
	}

	if uid := sess.GetValue("userid"); uid != 7 {
		t.Errorf("invalid user id returned expected '7' bot got '%d'", uid)
	}

	sess.RemoveKey("userid")

	if uid := sess.GetValue("userid"); uid != nil {
		t.Errorf("user id should be removed but got '%d'", uid)
	}
}
