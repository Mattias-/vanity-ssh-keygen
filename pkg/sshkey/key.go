package sshkey

type SSHKey interface {
	SSHPubkey() []byte
	SSHPrivkey() []byte
	New()
}
