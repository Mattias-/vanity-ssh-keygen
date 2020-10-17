package key

type SSHKey interface {
	SSHPubkey() ([]byte, error)
	SSHPrivkey() ([]byte, error)
}
