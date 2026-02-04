# Package: utils

**File:** `fileutil.go`

## Functions

### AtomicallySaveToFile

```go
func AtomicallySaveToFile(fileName string, data []byte) err error
```

AtomicallySaveToFile saves given data to the given file atomically
Appends a CRC64 checksum to the data before writing
File is either fully updated or not updated at all -> done by writing to a temporary file and renaming after

---

### CleanupFolders

```go
func CleanupFolders(dirs []string, ignoreFiles ...string) error
```

CleanupFolders ensures that each directory in dirs exists and removes all entries except in ignoreFiles

---

### LoadFromFile

```go
func LoadFromFile(path string) (data []byte, err error)
```

LoadFromFile loads data from the given file and verifies the checksum
Returns data w/o checksum

---

## Constants

### ChecksumSize

```go
const ChecksumSize = 8
```

ChecksumSize is the length (in bytes) of the CRC64 checksum header
