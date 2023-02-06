package worker

import (
	"github.com/Mattias-/vanity-ssh-keygen/pkg/keygen"
)

type Kgworker struct {
	results chan keygen.SSHKey
	count   uint64

	Matchfunc func(keygen.SSHKey) bool
	Keyfunc   func() keygen.SSHKey
}

func (w *Kgworker) Run() {
	k := w.Keyfunc()
	for {
		w.count += 1
		k.New()
		if w.Matchfunc(k) {
			// A result was found!
			break
		}
	}
	w.results <- k
}

func (w *Kgworker) Count() uint64 {
	return w.count
}

func (w *Kgworker) SetResultChan(results chan keygen.SSHKey) {
	w.results = results
}
