package databases

import "context"

type DocumentDatabase interface {
	// Database / schema
	CreateDatabase(ctx context.Context, projectID, id, name string) error
	GetDatabase(ctx context.Context, projectID, id string) (*Collection, error)
	ListDatabases(ctx context.Context, projectID string) ([]Collection, error)
	DeleteDatabase(ctx context.Context, projectID, id string) error

	// Collection
	CreateCollection(ctx context.Context, projectID, databaseID, collectionID, name string, attrs []Attribute, idxs []Index) error
	GetCollection(ctx context.Context, projectID, databaseID, collectionID string) (*Collection, error)
	ListCollections(ctx context.Context, projectID, databaseID string) ([]Collection, error)
	DeleteCollection(ctx context.Context, projectID, databaseID, collectionID string) error

	// Attribute / Index
	CreateAttribute(ctx context.Context, projectID, databaseID, collectionID string, attr Attribute) error
	CreateIndex(ctx context.Context, projectID, databaseID, collectionID string, idx Index) error

	// Document
	CreateDocument(ctx context.Context, projectID, databaseID, collectionID string, doc Document, perms []Permission) (Document, error)
	GetDocument(ctx context.Context, projectID, databaseID, collectionID, docID string) (*Document, error)
	UpdateDocument(ctx context.Context, projectID, databaseID, collectionID string, doc Document, perms []Permission) (Document, error)
	DeleteDocument(ctx context.Context, projectID, databaseID, collectionID, docID string) error
	ListDocuments(ctx context.Context, projectID, databaseID, collectionID string, q Query, roles []string) (*DocumentList, error)
	CountDocuments(ctx context.Context, projectID, databaseID, collectionID string, queries []string, roles []string) (int64, error)

	// System bootstrap
	EnsureSystemCollections(ctx context.Context, projectID string, internalID int64) error
}
