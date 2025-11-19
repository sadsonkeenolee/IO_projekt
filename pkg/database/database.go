package database

const (
	ItemNotFound string = "No information"
)

type DatasetFileMetadata struct {
	Directory string
	Type      string
	IsRead    bool
	CreatedAt *string
	ReadAt    *string
}

type Insertable interface {
	// Marker interface for a struct that is an Insertable item.
	Insertable()
}

type Queryable interface {
	// Marker interface for a struct that is a Queryable item.
	Queryable()
}
