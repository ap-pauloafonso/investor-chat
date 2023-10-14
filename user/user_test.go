package user

import "testing"

func Test_hashPassword(t *testing.T) {

	pass := "123"

	v1, _ := hashPassword(pass)

	v2, _ := hashPassword(pass)

	if !checkPasswordHash(pass, v1) {
		t.Errorf("aaaaaaaaaa")
	}

	if !checkPasswordHash(pass, v2) {
		t.Errorf("bbbbbbbbbbb")
	}

}
