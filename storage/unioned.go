package storage

import (
	"bytes"
	"fmt"
	"path"
	"strings"
)

type UnionedCache struct {
	cache map[string]record
	base  RCache
	temp  RWCache
}

func NewUnionedCache(fixtureDir string) *UnionedCache {
	return &UnionedCache{
		cache: map[string]record{},
		base:  NewFixtureStorage(fixtureDir),
		temp:  NewInMemoryCache(),
	}
}

func (u *UnionedCache) Clear() {
	u.cache = map[string]record{}
	u.base.Clear()
	u.temp.Clear()
}

func (u *UnionedCache) Reset() {
	u.cache = map[string]record{}
	u.temp.Clear()
}

func (u *UnionedCache) readFrom(link linkRecord) ([]byte, error) {
	switch link.layer {
	case 0:
		return u.base.Read(link.location)
	case 1:
		return u.temp.Read(link.location)
	default:
		return nil, fmt.Errorf("Read: %s: no such layer %d", link.location, link.layer)
	}
}

func (u *UnionedCache) Read(location string) (data []byte, err error) {
	if strings.HasSuffix(location, "/") {
		err = fmt.Errorf("Read: %s: location does not identify an object", location)
	} else if r, ok := u.cache[location]; ok {
		switch r.kind() {
		case kindHole:
			err = fmt.Errorf("Read: %s: no such record", location)
		case kindLink:
			data, err = u.readFrom(r.(linkRecord))
		default:
			err = fmt.Errorf("Read: %s: not a link or hole record", location)
		}
	} else if data, err = u.temp.Read(location); err == nil {
		u.cache[location] = linkRecord{layer: 1, location: location}
	} else if data, err = u.base.Read(location); err == nil {
		u.cache[location] = linkRecord{layer: 0, location: location}
	}
	return
}

func (u *UnionedCache) Exists(location string) bool {
	if r, ok := u.cache[location]; ok {
		return r.kind() == kindLink
	} else if u.temp.Exists(location) {
		u.cache[location] = linkRecord{layer: 1, location: location}
	} else if u.base.Exists(location) {
		u.cache[location] = linkRecord{layer: 0, location: location}
	} else {
		return false
	}
	return true
}

func (u *UnionedCache) List(location string) (subkeys []string, err error) {
	if !strings.HasSuffix(location, "/") {
		err = fmt.Errorf("List: %s: location does not identify a collection", location)
	} else if r, ok := u.cache[location]; ok {
		if r.kind() != kindCollection {
			err = fmt.Errorf("List: %s: not a prefix record", location)
		} else {
			subkeys = r.(collectionRecord).subkeys
		}
	} else {
		var basekeys, tempkeys []string

		if basekeys, err = u.base.List(location); err == nil {
			if tempkeys, err = u.temp.List(location); err == nil {
				subkeys = make([]string, 0, len(basekeys)+len(tempkeys))
				subkeys = append(subkeys, basekeys...)
				subkeys = append(subkeys, tempkeys...)
				u.cache[location] = collectionRecord{subkeys: subkeys}
			}
		}
	}
	return
}

func (u *UnionedCache) ReadList(location string) ([]byte, error) {
	if !strings.HasSuffix(location, "/") {
		return nil, fmt.Errorf("ReadList: %s: location does not identify a collection", location)
	}

	// Hydrate the file list if the data is not cached.
	var subkeys []string
	var err error
	if r, ok := u.cache[location]; !ok {
		subkeys, err = u.List(location)
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
		data, err := u.Read(subkey)
		if err != nil {
			return nil, err
		}
		buff.Write(data)
		delim = ","
	}
	buff.WriteString("]")

	// Cache the result and return it
	b := buff.Bytes()
	u.cache[location] = collectionRecord{data: b, subkeys: subkeys}
	return b, nil
}

func (u *UnionedCache) Write(location string, data []byte) (err error) {
	if strings.HasSuffix(location, "/") {
		err = fmt.Errorf("Write: %s: location does not identify an object", location)
	} else if err = u.temp.Write(location, data); err == nil {
		// Record the link
		if _, ok := u.cache[location]; !ok {
			u.cache[location] = linkRecord{layer: 1, location: location}
		}

		// Invalidate any cached list
		if parent := path.Dir(location); parent != "." {
			delete(u.cache, parent+"/")
		}
	}
	return
}

func (u *UnionedCache) Delete(location string) bool {
	if strings.HasSuffix(location, "/") || !u.Exists(location) {
		return false
	}

	var ok = true
	switch u.cache[location].(linkRecord).layer {
	case 0:
		u.cache[location] = holeRecord{}
	case 1:
		ok = u.temp.Delete(location)
		delete(u.cache, location)
	default:
		return false
	}

	// Invalidate any cached list
	if parent := path.Dir(location); parent != "." {
		delete(u.cache, parent+"/")
	}

	return ok
}
