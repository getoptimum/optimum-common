package identity_test

import (
	"testing"

	"github.com/getoptimum/optimum-common/pkg/identity"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
)

func TestPersistedIdentity(t *testing.T) {
	t.Run("default key type", func(t *testing.T) {
		dir := t.TempDir()
		_, err := identity.EnsureIdentity(dir)
		require.NoError(t, err)

		info, err := identity.ExtractIdentityFromDir(dir)
		require.NoError(t, err)
		t.Log(info.ID.String())

		key, err := crypto.UnmarshalPrivateKey(info.Key)
		require.NoError(t, err)
		id, err := peer.IDFromPrivateKey(key)
		require.NoError(t, err)
		require.Equal(t, id, info.ID)
	})
	t.Run("specified key type", func(t *testing.T) {
		dir := t.TempDir()
		_, err := identity.EnsureIdentity(dir, identity.GenIdentityEd25519)
		require.NoError(t, err)

		info, err := identity.ExtractIdentityFromDir(dir)
		require.NoError(t, err)
		t.Log(info.ID.String())

		key, err := crypto.UnmarshalPrivateKey(info.Key)
		require.NoError(t, err)
		id, err := peer.IDFromPrivateKey(key)
		require.NoError(t, err)
		require.Equal(t, id, info.ID)
	})
}
