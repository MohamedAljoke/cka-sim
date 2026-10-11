package password

import (
	"strings"
	"testing"
)

func TestVerifyAcceptsTheHashedPassword(t *testing.T) {
	encoded := mustHash(t, "correct horse")

	if !Verify(encoded, "correct horse") {
		t.Error("Verify rejected the right password")
	}
	if Verify(encoded, "wrong horse") {
		t.Error("Verify accepted a wrong password")
	}
}

func TestHashSaltsEachCall(t *testing.T) {
	first, second := mustHash(t, "same"), mustHash(t, "same")

	if first == second {
		t.Error("two hashes of one password are equal")
	}
}

func TestVerifyRejectsBadEncodings(t *testing.T) {
	good := mustHash(t, "pw")
	parts := strings.Split(good, "$")
	withParams := func(params string) string {
		p := append([]string(nil), parts...)
		p[3] = params
		return strings.Join(p, "$")
	}
	for name, encoded := range map[string]string{
		"empty":         "",
		"bcrypt":        "$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ01234",
		"bad salt":      strings.Replace(good, parts[4], "!!!", 1),
		"too much RAM":  withParams("m=1048576,t=2,p=1"),
		"too many runs": withParams("m=19456,t=100,p=1"),
		"too weak":      withParams("m=1024,t=1,p=1"),
	} {
		if Verify(encoded, "pw") {
			t.Errorf("%s: Verify accepted %q", name, encoded)
		}
	}
}

func TestVerifyDummyAlwaysFails(t *testing.T) {
	if VerifyDummy("cka-sim-dummy-verification-subject") {
		t.Error("VerifyDummy passed")
	}
}

func mustHash(t *testing.T, plain string) string {
	t.Helper()
	encoded, err := Hash(plain)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
