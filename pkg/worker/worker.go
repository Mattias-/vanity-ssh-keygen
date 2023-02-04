package worker

import (
	"github.com/Mattias-/vanity-ssh-keygen/pkg/sshkey"
)

type matcher interface {
	Match(*sshkey.SSHKey) bool
}

type keygen interface {
	New() sshkey.SSHKey
}

type Kgworker struct {
	Keygen  keygen
	Matcher matcher
	results chan sshkey.SSHKey
	count   uint64

	Matchfunc func(*sshkey.SSHKey) bool
	Keyfunc   func() sshkey.SSHKey
}

func (w *Kgworker) Run() {
	k := w.Keygen.New()
	for {
		w.count += 1
		k.New()
		if w.Matcher.Match(&k) {
			// A result was found!
			break
		}
	}
	w.results <- k
}

func (w *Kgworker) Count() uint64 {
	return w.count
}

func (w *Kgworker) SetResultChan(results chan sshkey.SSHKey) {
	w.results = results
}
