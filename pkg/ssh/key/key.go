package key

type SSHKey interface {
	SSHPubkey() []byte
	SSHPrivkey() []byte
	New()
}
