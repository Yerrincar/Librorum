package users

import "github.com/alexedwards/argon2id"

func (p *password) Set(plainText string) error {
	hash, err := argon2id.CreateHash(plainText, argon2id.DefaultParams)
	if err != nil {
		return err
	}

	p.Plaintext = &plainText
	p.Hash = hash
	return nil
}

func (p *password) Matches(plainText string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(plainText, p.Hash)
	if err != nil {
		return false, err
	}
	return match, nil
}
