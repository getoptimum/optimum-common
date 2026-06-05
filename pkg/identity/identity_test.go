package identity_test

import (
	"strings"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/identity"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
)

func TestPersistedIdentity(t *testing.T) {
	t.Run("default key type", func(t *testing.T) {
		dir := t.TempDir()
		pk, err := identity.EnsureIdentity(dir)
		require.NoError(t, err)
		require.EqualValues(t, crypto.Secp256k1, pk.Type())

		info, err := identity.ExtractIdentityFromDir(dir)
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(info.ID.String(), "16Uiu"))

		key, err := crypto.UnmarshalPrivateKey(info.Key)
		require.NoError(t, err)
		id, err := peer.IDFromPrivateKey(key)
		require.NoError(t, err)
		require.Equal(t, id, info.ID)
	})
	t.Run("specified key type", func(t *testing.T) {
		dir := t.TempDir()
		pk, err := identity.EnsureIdentity(dir, identity.GenIdentityEd25519)
		require.NoError(t, err)
		require.EqualValues(t, crypto.Ed25519, pk.Type())

		info, err := identity.ExtractIdentityFromDir(dir)
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(info.ID.String(), "12D3Koo"))

		key, err := crypto.UnmarshalPrivateKey(info.Key)
		require.NoError(t, err)
		id, err := peer.IDFromPrivateKey(key)
		require.NoError(t, err)
		require.Equal(t, id, info.ID)
	})
	t.Run("existing key is loaded not regenerated", func(t *testing.T) {
		dir := t.TempDir()
		_, err := identity.EnsureIdentity(dir, identity.GenerateIdentitySecp256k1)
		require.NoError(t, err)

		info, err := identity.ExtractIdentityFromDir(dir)
		require.NoError(t, err)

		pk, err := identity.EnsureIdentity(dir, identity.GenIdentityEd25519)
		require.NoError(t, err)
		require.EqualValues(t, crypto.Secp256k1, pk.Type())

		id, err := peer.IDFromPrivateKey(pk)
		require.NoError(t, err)
		require.Equal(t, info.ID, id)
	})
}
