package storage

type RCache interface {
	Clear()
	Exists(location string) bool
	Read(location string) ([]byte, error)
	ReadList(location string) ([]byte, error)
	List(location string) ([]string, error)
}

type RWCache interface {
	RCache
	Write(location string, object []byte) error
	Delete(location string) bool
}
