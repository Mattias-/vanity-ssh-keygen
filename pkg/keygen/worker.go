package keygen

type Worker struct {
	results chan SSHKey
	count   int64

	Matchfunc func(SSHKey) bool
	Keyfunc   func() SSHKey
}

func (w *Worker) Run() {
	k := w.Keyfunc()
	for {
		w.count += 1
		k.Generate()
		if w.Matchfunc(k) {
			// A result was found!
			break
		}
	}
	w.results <- k
}

func (w *Worker) Count() int64 {
	return w.count
}

func (w *Worker) SetResultChan(results chan SSHKey) {
	w.results = results
}
