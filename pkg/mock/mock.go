package mock

import (
	"github.com/williamlsh/vault/pkg/store"
)

type nopStore struct{}

// NewNopStore returns a store that does not do anything. It's especially
// useful in testing.
func NewNopStore() store.Store {
	return nopStore{}
}

func (m nopStore) KeepSecret(secret []byte) <-chan error {
	errc := make(chan error, 1)
	errc <- nil
	return errc
}
