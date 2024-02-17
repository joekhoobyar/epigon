package storage

import (
	"bytes"
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
		err = newError(location, ErrLocationNotObject)
	} else if r, ok := m.cache[location]; !ok {
		err = newError(location, ErrObjectNotFound)
	} else if r.kind() != kindObject {
		err = newError(location, ErrKindNotObject)
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
	if subdir, ok := strings.CutSuffix(location, "/"); !ok {
		err = newError(location, ErrLocationNotPrefix)
	} else if r, ok := m.cache[location]; ok {
		if r.kind() != kindCollection {
			err = newError(location, ErrKindNotPrefix)
		} else {
			subkeys = r.(collectionRecord).subkeys
		}
	} else {
		// Workaround since in-memory caches cannot test directory presence.
		// Test for parent key existence instead.
		parent := path.Dir(subdir)
		if parent != "." {
			if _, ok := m.cache[parent]; !ok {
				err = newError(location, ErrPrefixNotFound)
				return
			}
		}

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
		return nil, newError(location, ErrLocationNotPrefix)
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
		buff.Write(m.cache[subkey].(objectRecord).data)
	}
	buff.WriteString("]")

	// Cache the result and return it
	b := buff.Bytes()
	m.cache[location] = collectionRecord{data: b, subkeys: subkeys}
	return b, nil
}

func (m *InMemoryCache) Write(location string, data []byte) error {
	if strings.HasSuffix(location, "/") {
		return newError(location, ErrLocationNotObject)
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
