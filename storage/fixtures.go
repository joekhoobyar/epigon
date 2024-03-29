package storage

import (
	"bytes"
	"os"
	"path"
	"strings"
)

type FixtureStorage struct {
	Dir   string
	cache map[string]record
}

func NewFixtureStorage(dir string) *FixtureStorage {
	return &FixtureStorage{
		Dir:   dir,
		cache: map[string]record{},
	}
}

func (f *FixtureStorage) Clear() {
	f.cache = map[string]record{}
}

func (f *FixtureStorage) Reset() {
	// NO-OP: read-only storage
}

func (f *FixtureStorage) Read(location string) ([]byte, error) {
	if strings.HasSuffix(location, "/") {
		return nil, newError(location, ErrLocationNotObject)
	}

	if r, ok := f.cache[location]; ok {
		if r.kind() != kindObject {
			return nil, newError(location, ErrKindNotObject)
		}
		return r.(objectRecord).data, nil
	}

	path := path.Join(f.Dir, location) + ".json"
	buff, err := os.ReadFile(path)
	if err != nil {
		return nil, wrapFailure(err, location)
	}

	// cache the result and return it
	f.cache[location] = objectRecord{data: buff}
	return buff, nil
}

func (f *FixtureStorage) Exists(location string) bool {
	if r, ok := f.cache[location]; ok {
		return r.kind() == kindObject
	} else if strings.HasSuffix(location, "/") {
		return false
	} else {
		path := path.Join(f.Dir, location) + ".json"
		_, err := os.Stat(path)
		return err == nil
	}
}

func (f *FixtureStorage) List(location string) ([]string, error) {
	subdir, has := strings.CutSuffix(location, "/")
	if !has {
		return nil, newError(location, ErrLocationNotPrefix)
	}

	// Return the data if it is cached
	if r, ok := f.cache[location]; ok {
		if r.kind() != kindCollection {
			return nil, newError(location, ErrKindNotPrefix)
		}
		return r.(collectionRecord).subkeys, nil
	}

	// List files
	dir := path.Join(f.Dir, subdir)
	files, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, wrapError(err, location, ErrPrefixNotFound)
	} else if err != nil {
		return nil, wrapFailure(err, location)
	}

	// Cache the list of JSON files
	subkeys := make([]string, 0, len(files))
	for i := range files {
		name := files[i].Name()
		if basename, has := strings.CutSuffix(name, ".json"); has {
			subkeys = append(subkeys, path.Join(location, basename))
		}
	}

	// cache the result and return it
	f.cache[location] = collectionRecord{subkeys: subkeys}
	return subkeys, nil
}

func (f *FixtureStorage) ReadList(location string) ([]byte, error) {
	_, has := strings.CutSuffix(location, "/")
	if !has {
		return nil, newError(location, ErrLocationNotPrefix)
	}

	// Hydrate the file list if the data is not cached.
	var subkeys []string
	var err error
	if r, ok := f.cache[location]; !ok {
		subkeys, err = f.List(location)
		if err != nil {
			return nil, err
		}
	} else if r.kind() != kindCollection {
		return nil, newError(location, ErrKindNotPrefix)
	} else if r.(collectionRecord).data != nil {
		return r.(collectionRecord).data, nil
	} else {
		subkeys = r.(collectionRecord).subkeys
	}

	// Build an object record for this collection.
	buff := bytes.Buffer{}
	buff.WriteString("[")
	for i, subkey := range subkeys {
		if i > 0 {
			buff.WriteString(",")
		}
		data, err := f.Read(subkey)
		if err != nil {
			return nil, err
		}
		buff.Write(data)
	}
	buff.WriteString("]")

	// Cache the result and return it
	b := buff.Bytes()
	f.cache[location] = collectionRecord{data: b, subkeys: subkeys}
	return b, nil
}
