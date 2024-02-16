package storage

type recordKind uint8

const (
	kindNone = iota
	kindObject
	kindCollection
)

type objectRecord struct {
	data []byte
}

type collectionRecord struct {
	data    []byte
	subkeys []string
}

type record interface {
	kind() recordKind
}

func (objectRecord) kind() recordKind {
	return kindObject
}

func (collectionRecord) kind() recordKind {
	return kindCollection
}
