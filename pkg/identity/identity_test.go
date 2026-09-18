package identity_test

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/getoptimum/optimum-common/pkg/identity"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
)

const secp256k1PrivateKeySize = 32

func TestDeriveSecp256k1PrivateKey(t *testing.T) {
	t.Run("new generated key", func(t *testing.T) {
		// given
		pk, pkBytes := testPrivateKey(t)

		// when
		now := time.Now()
		seenGenerated := make(map[string]struct{})
		for i := 1; i < 10; i++ {
			newPk, err := identity.DeriveSecp256k1PrivateKey(pkBytes, now.Add(time.Duration(i)*time.Second).Format(time.DateTime))
			require.NoError(t, err)

			// then
			generated := testGetP2PKey(t, newPk).String()
			_, ok := seenGenerated[generated]
			require.False(t, ok)
			seenGenerated[generated] = struct{}{}
			require.NotEqual(t, testGetP2PKey(t, pk).String(), testGetP2PKey(t, newPk).String())
		}

		t.Run("should fallback to original", func(t *testing.T) {
			srcPK, errS := identity.DeriveSecp256k1PrivateKey(pkBytes, "")
			require.NoError(t, errS)
			require.Equal(t, testGetP2PKey(t, pk).String(), testGetP2PKey(t, srcPK).String())
		})
	})
	t.Run("well known keys", func(t *testing.T) {
		// given
		srcKey := "381d867840059044c8ff6a7ad825ce7c64d3e7cec72748ef5dd5bfbd880e6ee2"
		srcPkBytes, errS := hex.DecodeString(srcKey)
		require.NoError(t, errS)

		privateKey, err := crypto.UnmarshalSecp256k1PrivateKey(srcPkBytes)
		require.NoError(t, err)
		require.Equal(t, testGetP2PKey(t, privateKey).String(), "16Uiu2HAkxfNHQ2H71rNmu5nCvvVX6CTNMEDSbq4tK2SToxa5GZKv")

		// when
		table := map[string]string{
			"hello_world":      "16Uiu2HAkxtxWKfD1zDmpv6qzRdjXPcNs2HeTsrVZARHnGScWt5FF",
			"optimum_network":  "16Uiu2HAkxDHykqwsUW3Gpc8PTkCrtKw6dwCAFSyVH12RQM5KPK1v",
			"optimum_network1": "16Uiu2HAmNikQKQc3X5BmS4tgJsV5B2JQvRSaMkA2HSJxPdmvi2wJ",
		}
		for extraHash, expectedKey := range table {
			resD, errD := identity.DeriveSecp256k1PrivateKey(srcPkBytes, extraHash)
			require.NoError(t, errD)
			require.Equal(t, testGetP2PKey(t, resD).String(), expectedKey)
		}
	})
}

func TestDeriveSecp256k1PrivateKeyRejectsInvalidParent(t *testing.T) {
	curveOrder, err := hex.DecodeString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141")
	require.NoError(t, err)

	tests := map[string]struct {
		parentRaw []byte
		label     string
	}{
		"nil":              {label: "node"},
		"short":            {make([]byte, secp256k1PrivateKeySize-1), "node"},
		"long":             {make([]byte, secp256k1PrivateKeySize+1), "node"},
		"zero":             {make([]byte, secp256k1PrivateKeySize), "node"},
		"curve order":      {curveOrder, "node"},
		"invalid fallback": {make([]byte, secp256k1PrivateKeySize), ""},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err = identity.DeriveSecp256k1PrivateKey(tt.parentRaw, tt.label)
			require.Error(t, err)
		})
	}
}

func testGetP2PKey(t *testing.T, key crypto.PrivKey) peer.ID {
	t.Helper()

	id, err := peer.IDFromPrivateKey(key)
	require.NoError(t, err)
	return id
}

func testPrivateKey(t *testing.T) (key crypto.PrivKey, privBytes []byte) {
	t.Helper()
	var err error
	key, _, err = crypto.GenerateSecp256k1Key(rand.Reader)
	require.NoError(t, err)

	privBytes, err = key.Raw()
	require.NoError(t, err)
	require.Len(t, privBytes, secp256k1PrivateKeySize)
	return key, privBytes
}

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
