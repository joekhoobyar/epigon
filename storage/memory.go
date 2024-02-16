package storage

import (
	"bytes"
	"fmt"
	"path"
	"slices"
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

func (m *InMemoryCache) Clear() {
	m.cache = map[string]record{}
}

func (m *InMemoryCache) Reset() {
	m.cache = map[string]record{}
}

func (m *InMemoryCache) Read(location string) (data []byte, err error) {
	if strings.HasSuffix(location, "/") {
		err = fmt.Errorf("Read: %s: location does not identify an object", location)
	} else if r, ok := m.cache[location]; !ok {
		err = fmt.Errorf("Read: %s: no such record", location)
	} else if r.kind() != kindObject {
		err = fmt.Errorf("Read: %s: not an object record", location)
	} else {
		data = r.(objectRecord).data
	}
	return
}

func (m *InMemoryCache) Exists(location string) bool {
	r, ok := m.cache[location]
	return ok && r.kind() == kindObject
}

func (m *InMemoryCache) List(location string) (subkeys []string, err error) {
	if !strings.HasSuffix(location, "/") {
		err = fmt.Errorf("List: %s: location does not identify a collection", location)
	} else if r, ok := m.cache[location]; ok {
		if r.kind() != kindCollection {
			err = fmt.Errorf("List: %s: not a prefix record", location)
		} else {
			subkeys = r.(collectionRecord).subkeys
		}
	} else {
		subkeys = make([]string, 0, len(m.cache))
		for key := range m.cache {
			if strings.HasPrefix(key, location) {
				subkeys = append(subkeys, key)
			}
		}
		slices.Sort(subkeys)
		m.cache[location] = collectionRecord{subkeys: subkeys}
	}
	return
}

func (m *InMemoryCache) ReadList(location string) ([]byte, error) {
	if !strings.HasSuffix(location, "/") {
		return nil, fmt.Errorf("ReadList: %s: location does not identify a collection", location)
	}

	// Hydrate the file list if the data is not cached.
	var subkeys []string
	var err error
	if r, ok := m.cache[location]; !ok {
		subkeys, err = m.List(location)
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
		buff.Write(m.cache[subkey].(objectRecord).data)
		delim = ","
	}
	buff.WriteString("]")

	// Cache the result and return it
	b := buff.Bytes()
	m.cache[location] = collectionRecord{data: b, subkeys: subkeys}
	return b, nil
}

func (m *InMemoryCache) Write(location string, data []byte) error {
	if strings.HasSuffix(location, "/") {
		return fmt.Errorf("Write: %s: location does not identify an object", location)
	}

	m.cache[location] = objectRecord{data: data}

	// Invalidate any cached list
	if parent := path.Dir(location); parent != "." {
		delete(m.cache, parent+"/")
	}
	return nil
}

func (m *InMemoryCache) Delete(location string) bool {
	if strings.HasSuffix(location, "/") {
		return false
	}

	// Record whether it exists, then delete it
	_, ok := m.cache[location]
	delete(m.cache, location)

	// Invalidate any cached list
	if parent := path.Dir(location); parent != "." {
		delete(m.cache, parent+"/")
	}
	return ok
}
