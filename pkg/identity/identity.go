package identity

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/getoptimum/optimum-common/pkg/utils"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

const KeyFilename = "p2p.key"

type IdentityInfo struct {
	Key []byte
	ID  peer.ID // this is needed only to simplify integration with some testing tools
}

type Option func() (crypto.PrivKey, error)

// EnsureIdentity generates an identity key file in given directory.
func EnsureIdentity(dir string, opt ...Option) (crypto.PrivKey, error) {
	if len(opt) > 0 && opt[0] != nil {
		return ensureIdentity(dir, opt[0])
	}
	return ensureIdentity(dir, GenIdentityEd25519)
}

// GenerateIdentitySecp256k1 creates a new random Secp256k1 private key.
func GenerateIdentitySecp256k1() (crypto.PrivKey, error) {
	pk, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate secp256k1 identity: %w", err)
	}
	return pk, nil
}

// GenIdentityEd25519 creates a new random Ed25519 private key.
func GenIdentityEd25519() (crypto.PrivKey, error) {
	pk, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate ed25519 identity: %w", err)
	}
	return pk, nil
}

func ExtractIdentityFromDir(dir string) (*IdentityInfo, error) {
	path := filepath.Join(dir, KeyFilename)
	data, err := utils.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}
	var info IdentityInfo
	err = json.Unmarshal(data, &info)
	if err != nil {
		return nil, fmt.Errorf("unmarshal file content from %s into %+v: %w", path, info, err)
	}
	return &info, nil
}

// ensureIdentity is an internal helper that generates or loads an identity key file in the given directory using the provided key generation function.
func ensureIdentity(dir string, keygen func() (crypto.PrivKey, error)) (crypto.PrivKey, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil { //nolint:gosec // directory permissions
		return nil, fmt.Errorf("ensure that directory %s exist: %w", dir, err)
	}
	info, err := ExtractIdentityFromDir(dir)
	if err == nil {
		pk, err := crypto.UnmarshalPrivateKey(info.Key)
		if err != nil {
			return nil, fmt.Errorf("unmarshal privkey: %w", err)
		}
		return pk, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		key, err := keygen()
		if err != nil {
			return nil, err
		}
		id, err := peer.IDFromPrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("generated key is malformed: %w", err)
		}
		raw, err := crypto.MarshalPrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("marshal private key: %w", err)
		}
		data, err := json.Marshal(IdentityInfo{
			Key: raw,
			ID:  id,
		})
		if err != nil {
			return nil, err
		}
		if err = utils.AtomicallySaveToFile(filepath.Join(dir, KeyFilename), data); err != nil {
			return nil, fmt.Errorf("write identity data: %w", err)
		}
		return key, nil
	}
	return nil, fmt.Errorf("read key from disk: %w", err)
}
