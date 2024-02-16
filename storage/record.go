package storage

type recordKind uint8

const (
	kindNone = iota
	kindObject
	kindCollection
	kindLink
	kindHole
)

type objectRecord struct {
	data []byte
}

type collectionRecord struct {
	data    []byte
	subkeys []string
}

type linkRecord struct {
	layer    int
	location string
}

type holeRecord struct{}

type record interface {
	kind() recordKind
}

func (objectRecord) kind() recordKind {
	return kindObject
}

func (collectionRecord) kind() recordKind {
	return kindCollection
}

func (linkRecord) kind() recordKind {
	return kindLink
}

func (holeRecord) kind() recordKind {
	return kindHole
}
