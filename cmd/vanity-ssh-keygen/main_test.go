package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Mattias-/vanity-ssh-keygen/pkg/keygen"
)

type mockKey struct {
	pub  []byte
	priv []byte
}

func (m *mockKey) SSHPubkey() []byte  { return m.pub }
func (m *mockKey) SSHPrivkey() []byte { return m.priv }
func (m *mockKey) Generate()          {}

func TestVersionString(t *testing.T) {
	v := versionString()
	if !strings.Contains(v, "dev") {
		t.Errorf("Expected version string to contain 'dev', got %s", v)
	}
}

func TestOutputPEM(t *testing.T) {
	tmpDir := t.TempDir()
	a := &app{
		config: config{
			MatchString: "test",
			OutputDir:   tmpDir,
		},
	}

	key := &mockKey{
		pub:  []byte("test-pub"),
		priv: []byte("test-priv"),
	}

	a.outputPEM(1*time.Second, key)

	privFile := filepath.Join(tmpDir, "test")
	pubFile := filepath.Join(tmpDir, "test.pub")

	if _, err := os.Stat(privFile); os.IsNotExist(err) {
		t.Errorf("Private key file not created")
	}
	if _, err := os.Stat(pubFile); os.IsNotExist(err) {
		t.Errorf("Public key file not created")
	}

	//nolint:gosec // G304: Path is controlled by the test via t.TempDir()
	content, _ := os.ReadFile(privFile)
	if string(content) != "test-priv" {
		t.Errorf("Unexpected private key content: %s", string(content))
	}
}

func TestOutputJSON(t *testing.T) {
	tmpDir := t.TempDir()
	a := &app{
		config: config{
			MatchString: "test",
			OutputDir:   tmpDir,
		},
	}

	key := &mockKey{
		pub:  []byte("test-pub"),
		priv: []byte("test-priv"),
	}

	a.outputJSON(1*time.Second, key)

	jsonFile := filepath.Join(tmpDir, "result.json")

	if _, err := os.Stat(jsonFile); os.IsNotExist(err) {
		t.Errorf("JSON file not created")
	}

	//nolint:gosec // G304: Path is controlled by the test via t.TempDir()
	content, _ := os.ReadFile(jsonFile)
	var out OutputData
	if err := json.Unmarshal(content, &out); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if out.PublicKey != "test-pub" || out.PrivateKey != "test-priv" || out.Metadata.FindString != "test" {
		t.Errorf("Unexpected JSON content: %+v", out)
	}
}

func TestRunKeygen(t *testing.T) {
	a := &app{
		config: config{
			Threads:          1,
			StatsLogInterval: 0,
		},
	}

	mockM := &mockMatcher{match: true}
	mockK := func() keygen.SSHKey {
		return &mockKey{pub: []byte("match"), priv: []byte("priv")}
	}

	var capturedResult keygen.SSHKey
	outputter := func(elapsed time.Duration, result keygen.SSHKey) {
		capturedResult = result
	}

	a.runKeygen(mockM, mockK, outputter)

	if capturedResult == nil {
		t.Fatal("Result not captured")
	}
	if string(capturedResult.SSHPubkey()) != "match" {
		t.Errorf("Expected 'match', got %s", string(capturedResult.SSHPubkey()))
	}
}

type mockMatcher struct {
	match bool
}

func (m *mockMatcher) SetMatchString(s string)    {}
func (m *mockMatcher) Match(k keygen.SSHKey) bool { return m.match }
