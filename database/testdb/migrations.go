package testdb

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

// migrationFiles walks the embedded migrations FS, filters by suffix, sorts by
// filename (which starts with a zero-padded version number), and returns the
// sorted filenames and their contents.
func migrationFiles(migrations fs.FS, suffix string) ([]string, []string, error) {
	entries, err := fs.ReadDir(migrations, "migrations")
	if err != nil {
		return nil, nil, fmt.Errorf("reading migrations directory: %w", err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), suffix) {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	contents := make([]string, len(names))
	for i, name := range names {
		data, err := fs.ReadFile(migrations, "migrations/"+name)
		if err != nil {
			return nil, nil, fmt.Errorf("reading migration %s: %w", name, err)
		}
		contents[i] = string(data)
	}

	return names, contents, nil
}

// migrationHash computes a SHA256 hash of migration filenames and contents.
// The hash is content-addressed: it changes when any migration is added or
// modified, which automatically invalidates caches keyed by the hash.
func migrationHash(names, contents []string) string {
	h := sha256.New()
	for i := range names {
		io.WriteString(h, names[i])
		io.WriteString(h, contents[i])
	}
	return hex.EncodeToString(h.Sum(nil))
}

// migrationVersion extracts the version number from the last migration filename.
// Filenames follow the pattern NNN_description.up.{suffix}.sql
func migrationVersion(names []string) (int, error) {
	if len(names) == 0 {
		return 0, nil
	}
	parts := strings.SplitN(names[len(names)-1], "_", 2)
	v, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("parsing migration version from %s: %w", names[len(names)-1], err)
	}
	return v, nil
}
