package keygen

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestSshAdd(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}
	for _, v := range KeygenList() {
		SSHAddCompatible(t, v.f())
	}
}

func SSHAddCompatible(t *testing.T, k SSHKey) {
	k.New()
	pk := k.SSHPrivkey()

	keyfile := t.TempDir() + "/k"
	err := os.WriteFile(keyfile, pk, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	{
		out, err := exec.Command("ssh-add", "-t", "1", keyfile).CombinedOutput()
		t.Logf("%s", out)
		var eerr *exec.ExitError
		if errors.As(err, &eerr) {
			t.Logf("Exit code: %d", eerr.ExitCode())
		}
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(string(out), "Identity added:") {
			t.Fail()
		}
	}
}
