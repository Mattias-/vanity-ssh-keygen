package keygen

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestSshAdd(t *testing.T) {
	t.Skip()
	for _, v := range KeygenList() {
		SSHAddCompatible(t, v())
	}
}

func SSHAddCompatible(t *testing.T, k SSHKey) {
	t.Helper()

	k.New()
	pk := k.SSHPrivkey()

	keyfile := t.TempDir() + "/k"
	err := os.WriteFile(keyfile, pk, 0600)
	if err != nil {
		t.Fatal(err)
	}

	out, err := exec.Command("ssh-add", keyfile).CombinedOutput()
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
