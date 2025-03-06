package keygen

type SSHKey interface {
	SSHPubkey() []byte
	SSHPrivkey() []byte
	Generate()
}

type Keygen func() SSHKey

var keygens = map[string]Keygen{}

func RegisterKeygen(name string, k Keygen) {
	keygens[name] = k
}

func Names() []string {
	keys := make([]string, 0, len(keygens))
	for k := range keygens {
		keys = append(keys, k)
	}
	return keys
}

func Get(name string) (Keygen, bool) {
	k, ok := keygens[name]
	return k, ok
}
