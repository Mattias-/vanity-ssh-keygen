package keygen

type SSHKey interface {
	SSHPubkey() []byte
	SSHPrivkey() []byte
	Generate()
}

type Keygen func() SSHKey

type namedKeygen struct {
	name   string
	keygen Keygen
}

var keygens = []namedKeygen{}

func RegisterKeygen(name string, k Keygen) {
	keygens = append(keygens, namedKeygen{name, k})
}

func Names() []string {
	names := make([]string, 0, len(keygens))
	for _, k := range keygens {
		names = append(names, k.name)
	}
	return names
}

func Get(name string) (Keygen, bool) {
	for _, k := range keygens {
		if k.name == name {
			return k.keygen, true
		}
	}
	return nil, false
}
