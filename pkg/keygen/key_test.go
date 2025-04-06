package keygen

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/keygen/ed25519"
	"github.com/Mattias-/vanity-ssh-keygen/pkg/keygen/rsa"
)

func TestSshAdd_RSA2048(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}
	SSHAddCompatible(t, rsa.New(2048))
}

func TestSshAdd_RSA4096(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}
	SSHAddCompatible(t, rsa.New(4096))
}

func TestSshAdd_ED25519(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}
	SSHAddCompatible(t, ed25519.New())
}

func SSHAddCompatible(t *testing.T, k SSHKey) {
	k.Generate()
	pk := k.SSHPrivkey()

	keyfile := t.TempDir() + "/k"
	err := os.WriteFile(keyfile, pk, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	{
		// #nosec G204
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
