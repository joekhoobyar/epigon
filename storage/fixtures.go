package storage

import (
	"bytes"
	"fmt"
	"log"
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

func (f *FixtureStorage) Read(location string) ([]byte, error) {
	if strings.HasSuffix(location, "/") {
		return nil, fmt.Errorf("Read: %s: location does not identify an object", location)
	}

	if r, ok := f.cache[location]; ok {
		if r.kind() != kindObject {
			return nil, fmt.Errorf("Read: %s: not an object record", location)
		}
		return r.(objectRecord).data, nil
	}

	path := path.Join(f.Dir, location) + ".json"
	buff, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Read: %s: readfile: %s", location, err)
	}

	// cache the result and return it
	f.cache[location] = objectRecord{data: buff}
	return buff, nil
}

func (f *FixtureStorage) Exists(location string) bool {
	r, ok := f.cache[location]
	return ok && r.kind() == kindObject
}

func (f *FixtureStorage) List(location string) ([]string, error) {
	subdir, has := strings.CutSuffix(location, "/")
	if !has {
		return nil, fmt.Errorf("List: %s: location does not identify a collection", location)
	}

	// Return the data if it is cached
	if r, ok := f.cache[location]; ok {
		if r.kind() != kindCollection {
			return nil, fmt.Errorf("List: %s: not a prefix record", location)
		}
		return r.(collectionRecord).subkeys, nil
	}

	// List files
	dir := path.Join(f.Dir, subdir)
	files, err := os.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		log.Panicf("readdir: %s: %s", location, err)
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
		return nil, fmt.Errorf("ReadList: %s: location does not identify a collection", location)
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
		return nil, fmt.Errorf("ReadList: %s: not a prefix record", location)
	} else if r.(collectionRecord).data != nil {
		return r.(collectionRecord).data, nil
	} else {
		subkeys = r.(collectionRecord).subkeys
	}

	// Build an object record for this collection.
	buff := bytes.Buffer{}
	delim := "["
	for _, subkey := range subkeys {
		buff.WriteString(delim)
		data, err := f.Read(subkey)
		if err != nil {
			return nil, err
		}
		buff.Write(data)
		delim = ","
	}
	buff.WriteString("]")

	// Cache the result and return it
	b := buff.Bytes()
	f.cache[location] = collectionRecord{data: b, subkeys: subkeys}
	return b, nil
}
