package utils

import (
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRWMapBasicOperations(t *testing.T) {
	t.Parallel()

	for name, ctor := range map[string]func() *RWMap[int, string]{
		"empty":    func() *RWMap[int, string] { return NewRWMap[int, string]() },
		"from map": func() *RWMap[int, string] { return NewRWMapFromStdMap(make(map[int]string)) },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rwMap := ctor()
			require.Empty(t, rwMap.LoadAll())

			data := make(map[int]string)
			var wg sync.WaitGroup
			var mu sync.Mutex
			var errs []error

			for i := 0; i < 100; i++ {
				id := uuid.NewString() // generate and capture safely outside goroutine
				data[i] = id
				wg.Add(1)

				go func(i int, expected string) {
					defer wg.Done()

					rwMap.Store(i, expected)
					value, ok := rwMap.Load(i)
					if !ok || value != expected {
						mu.Lock()
						errs = append(errs, fmt.Errorf("load mismatch for %d: %q %v", i, value, ok))
						mu.Unlock()
					}

					rwMap.Delete(i)
					value, ok = rwMap.Load(i)
					if ok || value != "" {
						mu.Lock()
						errs = append(errs, fmt.Errorf("expected delete for %d, got %q %v", i, value, ok))
						mu.Unlock()
					}

					rwMap.Store(i, expected)
				}(i, id)
			}
			wg.Wait()
			require.Empty(t, errs)

			require.Equal(t, data, rwMap.LoadAll())

			keys := rwMap.Keys()
			require.Len(t, keys, len(data))

			seen := make(map[int]struct{}, len(keys))
			rwMap.Range(func(key int, value string) bool {
				require.Equal(t, data[key], value)
				seen[key] = struct{}{}
				return true
			})
			require.Equal(t, len(data), len(seen))

			applied := uuid.NewString()
			rwMap.DoAndApply(1, func(string) string { return applied })
			val, ok := rwMap.Load(1)
			require.True(t, ok)
			require.Equal(t, applied, val)

			called := false
			ok = rwMap.Do(1, func(string) { called = true })
			require.True(t, called)
			require.True(t, ok)

			rwMap.DeleteAll()
			require.Zero(t, rwMap.Len())
			require.Empty(t, rwMap.LoadAll())
		})
	}
}

func TestRWMapReplaceAndLoadAllAndErase(t *testing.T) {
	rwMap := NewRWMap[int, string]()
	rwMap.Store(1, "one")
	rwMap.Store(2, "two")

	replaced := rwMap.Replace(map[int]string{3: "three"})
	require.Equal(t, map[int]string{1: "one", 2: "two"}, replaced)
	require.Equal(t, map[int]string{3: "three"}, rwMap.LoadAll())

	collected := rwMap.LoadAllAndErase()
	require.Equal(t, map[int]string{3: "three"}, collected)
	require.Empty(t, rwMap.LoadAll())
}

func TestRWMapConcurrentReplace(t *testing.T) {
	t.Parallel()

	rwMap := NewRWMap[int, string]()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			rwMap.Replace(map[int]string{i: uuid.NewString()})
		}(i)
	}
	wg.Wait()

	require.Equal(t, 1, rwMap.Len())
}

const (
	numGoroutines = 200
	numFunctions  = 200
	numOperations = 1000
)

func TestRWMapConcurrent(t *testing.T) {
	rwMap := NewRWMap[int, int]()
	var wg sync.WaitGroup

	funcsList := []func(i, j int){
		func(i, j int) {
			rwMap.Store(i, j)
		},
		func(i, j int) {
			rwMap.Load(i)
		},
		func(i, j int) {
			rwMap.Delete(i)
		},
		func(_, _ int) {
			rwMap.LoadAll()
		},
		func(_, _ int) {
			rwMap.Keys()
		},
		func(_, _ int) {
			rwMap.Len()
		},
		func(_, _ int) {
			rwMap.Replace(map[int]int{13: 13, 21: 21, 34: 34})
		},
		func(i, j int) {
			rwMap.DoAndApply(i, func(val int) int {
				return val + j
			})
		},
		func(i, _ int) {
			rwMap.Do(i, func(val int) {
				_ = val
			})
		},
		func(_, _ int) {
			rwMap.LoadAllAndErase()
		},
		func(_, _ int) {
			rwMap.DeleteAll()
		},
		func(_, _ int) {
			rwMap.Range(func(key, value int) bool {
				_ = key
				_ = value
				return true
			})
		},
	}

	for i := 0; i < numGoroutines; i++ {
		for _, f := range funcsList {
			wg.Add(1)
			go func(i int, f func(i, j int)) {
				defer wg.Done()
				wrapper(func() {
					f(i, i)
				})
			}(i, f)
		}
	}
	wg.Wait()
}

func wrapper(caller func()) {
	var wg sync.WaitGroup
	for i := 0; i < numFunctions; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				caller()
			}
		}(i)
	}
	wg.Wait()
}
