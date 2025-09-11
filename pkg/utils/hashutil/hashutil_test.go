package hashutil_test

import (
	"testing"

	"github.com/getoptimum/optimum-common/pkg/utils/hashutil"
	"github.com/stretchr/testify/require"
)

func TestHashSHA512(t *testing.T) {
	table := map[string]string{
		"Hello, World!": "374d794a95cdcfd8b35993185fef9ba368f160d8daf432d08ba9f1ed1e5abe6cc69291e0fa2fe0006a52570ef18c19def4e617c33ce52ef0a6e5fbe318cb0387",
		"Hello, 123!":   "c242f458c7473bb510c3d273e593f4d16ea4d45a1b763308ae4e780c5e6175afdfc3b2735ea81a9a9b84a8b5515749d5d443f641d7aed13236295303ecf420b5",
		"Hello, 123":    "84df6bdafdaa325beeaa4dedf46e6519e350ac7c9936d44d5b1de84359572d3d7047bece9f25dbb12876b9f307bb994f0df737b87757a0081583f3b23b7d4a4b",
	}
	for src, res := range table {
		require.Equal(t, res, hashutil.HashSHA512([]byte(src)))
	}
}

func TestHashSHA256(t *testing.T) {
	table := map[string]string{
		"Hello, World!": "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f",
		"Hello, 123!":   "52ce7f8d6d1f955e01b896c8fb38de421b4f9d0a2978fb1e3a3c9f3a6efa80ff",
		"Hello, 123":    "30b6bfae65bce9ae9ab1cef925407ddc3bcc3ee3ccbb4991619a4d7cd0c72675",
	}
	for src, res := range table {
		require.Equal(t, res, hashutil.HashSHA256([]byte(src)))
	}
}

func BenchmarkHashXXHash(b *testing.B) {
	data := []byte("Hello, World!")
	for i := 0; i < b.N; i++ {
		hashutil.HashXXHash(data)
	}
}

func BenchmarkHashSHA256String(b *testing.B) {
	data := []byte("Hello, World!")
	for i := 0; i < b.N; i++ {
		hashutil.HashSHA256String(data)
	}
}
