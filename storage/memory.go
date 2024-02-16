package storage

import (
	"bytes"
	"fmt"
	"path"
	"strings"
)

type InMemoryCache struct {
	cache map[string]record
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		cache: map[string]record{},
	}
}

func (f *InMemoryCache) Clear() {
	f.cache = map[string]record{}
}

func (f *InMemoryCache) Read(location string) (data []byte, err error) {
	if strings.HasSuffix(location, "/") {
		err = fmt.Errorf("Read: %s: location does not identify an object", location)
	} else if r, ok := f.cache[location]; !ok {
		err = fmt.Errorf("Read: %s: no such record", location)
	} else if r.kind() != kindObject {
		err = fmt.Errorf("Read: %s: not an object record", location)
	} else {
		data = r.(objectRecord).data
	}
	return
}

func (f *InMemoryCache) Exists(location string) bool {
	r, ok := f.cache[location]
	return ok && r.kind() == kindObject
}

func (f *InMemoryCache) List(location string) (subkeys []string, err error) {
	if !strings.HasSuffix(location, "/") {
		err = fmt.Errorf("List: %s: location does not identify a collection", location)
	} else if r, ok := f.cache[location]; ok {
		if r.kind() != kindCollection {
			err = fmt.Errorf("List: %s: not a prefix record", location)
		} else {
			subkeys = r.(collectionRecord).subkeys
		}
	} else {
		subkeys = make([]string, 0, len(f.cache))
		for key := range f.cache {
			if strings.HasPrefix(key, location) {
				subkeys = append(subkeys, key)
			}
		}
		f.cache[location] = collectionRecord{subkeys: subkeys}
	}
	return
}

func (f *InMemoryCache) ReadList(location string) ([]byte, error) {
	if !strings.HasSuffix(location, "/") {
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
		buff.Write(f.cache[subkey].(objectRecord).data)
		delim = ","
	}
	buff.WriteString("]")

	// Cache the result and return it
	b := buff.Bytes()
	f.cache[location] = collectionRecord{data: b, subkeys: subkeys}
	return b, nil
}

func (f *InMemoryCache) Write(location string, data []byte) error {
	if strings.HasSuffix(location, "/") {
		return fmt.Errorf("Write: %s: location does not identify an object", location)
	}

	f.cache[location] = objectRecord{data: data}

	// Invalidate any cached list
	if parent := path.Dir(location); parent != "." {
		delete(f.cache, parent+"/")
	}
	return nil
}

func (f *InMemoryCache) Delete(location string) bool {
	if strings.HasSuffix(location, "/") {
		return false
	}

	// Record whether it exists, then delete it
	_, ok := f.cache[location]
	delete(f.cache, location)

	// Invalidate any cached list
	if parent := path.Dir(location); parent != "." {
		delete(f.cache, parent+"/")
	}
	return ok
}
