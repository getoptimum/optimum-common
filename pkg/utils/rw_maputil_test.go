package utils_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRWMapBasicOperations(t *testing.T) {
	t.Parallel()

	for name, ctor := range map[string]func() *utils.RWMap[int, string]{
		"empty":    func() *utils.RWMap[int, string] { return utils.NewRWMap[int, string]() },
		"from map": func() *utils.RWMap[int, string] { return utils.NewRWMapFromStdMap(make(map[int]string)) },
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
	rwMap := utils.NewRWMap[int, string]()
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

	rwMap := utils.NewRWMap[int, string]()
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
	rwMap := utils.NewRWMap[int, int]()
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

func Test_RWMap_DoAndApply_Concurrent(t *testing.T) {
	m := utils.NewRWMap[string, int]()

	const goroutines = 50
	const keys = 10

	var wg sync.WaitGroup
	start := make(chan struct{})

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			<-start
			key := fmt.Sprintf("key-%d", i%keys)
			m.DoAndApply(key, func(v int) int { return v + 1 })
		}(i)
	}

	close(start)
	wg.Wait()

	// Validate results
	total := 0
	for i := 0; i < keys; i++ {
		val, ok := m.Load(fmt.Sprintf("key-%d", i))
		require.True(t, ok)
		total += val
	}
	require.Equal(t, goroutines, total)
}

func Test_RWMap_LoadAll_ConcurrentWithWrites(t *testing.T) {
	m := utils.NewRWMap[int, string]()
	var wg sync.WaitGroup
	const ops = 100

	for i := 0; i < ops; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m.Store(i, uuid.NewString())
		}(i)
	}

	for i := 0; i < ops; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = m.LoadAll() // should not panic even while writes happen
		}()
	}

	wg.Wait()
	require.GreaterOrEqual(t, len(m.LoadAll()), 0)
}

func Test_RWMap_DeleteAll_ConcurrentSafety(t *testing.T) {
	m := utils.NewRWMap[int, string]()
	for i := 0; i < 100; i++ {
		m.Store(i, uuid.NewString())
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%10 == 0 {
				m.DeleteAll()
			} else {
				m.Store(i, uuid.NewString())
			}
			_ = m.LoadAll()
		}(i)
	}
	wg.Wait()

	// Ensure map is still valid and not nil
	all := m.LoadAll()
	require.NotNil(t, all)
}

func Test_RWMap_LoadAllAndErase(t *testing.T) {
	m := utils.NewRWMap[string, string]()
	m.Store("a", "A")
	m.Store("b", "B")

	erasedCopy := m.LoadAllAndErase()
	require.Equal(t, map[string]string{"a": "A", "b": "B"}, erasedCopy)
	require.Equal(t, map[string]string{}, m.LoadAll(), "map should be empty after erase")
}
