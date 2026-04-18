package keygen

import "context"

type Worker struct {
	results chan SSHKey
	count   int64

	Matchfunc func(SSHKey) bool
	Keyfunc   func() SSHKey
}

func (w *Worker) Run(ctx context.Context) {
	k := w.Keyfunc()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		w.count += 1
		k.Generate()
		if w.Matchfunc(k) {
			// A result was found!
			break
		}
	}
	select {
	case w.results <- k:
	case <-ctx.Done():
	}
}

func (w *Worker) Count() int64 {
	return w.count
}

func (w *Worker) SetResultChan(results chan SSHKey) {
	w.results = results
}
