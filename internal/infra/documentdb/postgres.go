package documentdb

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/deeploop-ai/fleet/internal/domain/databases"
	"github.com/deeploop-ai/fleet/internal/infra/bun/model"
	"github.com/deeploop-ai/fleet/internal/infra/clients"
	"github.com/deeploop-ai/fleet/pkg/crud"
	"github.com/deeploop-ai/fleet/pkg/idgen"
	"github.com/deeploop-ai/fleet/pkg/query"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

var safeNameRe = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
var docIDRe = regexp.MustCompile(`^[a-zA-Z0-9_.:-]{1,64}$`)

const maxQueryLimit = 100

// ErrPermissionDenied is returned when the caller lacks document-level permission.
var ErrPermissionDenied = errors.New("permission denied")

type postgresDocumentDB struct {
	db *clients.Database
}

func NewPostgresDocumentDatabase(db *clients.Database) databases.DocumentDB {
	return &postgresDocumentDB{db: db}
}

func (p *postgresDocumentDB) conn(ctx context.Context) bun.IDB {
	return p.db.Conn(ctx)
}

func (p *postgresDocumentDB) CreateDatabase(ctx context.Context, projectID, id, name string) error {
	internalID, err := p.resolveInternalID(ctx, projectID)
	if err != nil {
		return err
	}
	schema := schemaName(internalID, id)
	if _, err := p.db.DB.ExecContext(ctx, fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, quoteIdent(schema))); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}
	if err := p.ensurePermsTable(ctx, schema); err != nil {
		return err
	}
	m := &model.DocumentDatabase{
		ID:        id,
		ProjectID: projectID,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err = p.db.NewInsert().Model(m).Exec(ctx)
	return err
}

func (p *postgresDocumentDB) GetDatabase(ctx context.Context, projectID, id string) (*databases.Collection, error) {
	m := new(model.DocumentDatabase)
	err := p.db.NewSelect().Model(m).Where("project_id = ? AND id = ?", projectID, id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &databases.Collection{
		ID:        m.ID,
		ProjectID: m.ProjectID,
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}, nil
}

func (p *postgresDocumentDB) ListDatabases(ctx context.Context, projectID string) ([]databases.Collection, error) {
	var ms []model.DocumentDatabase
	err := p.db.NewSelect().Model(&ms).Where("project_id = ?", projectID).Order("created_at DESC").Scan(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]databases.Collection, len(ms))
	for i := range ms {
		out[i] = databases.Collection{
			ID:        ms[i].ID,
			ProjectID: ms[i].ProjectID,
			Name:      ms[i].Name,
			CreatedAt: ms[i].CreatedAt,
			UpdatedAt: ms[i].UpdatedAt,
		}
	}
	return out, nil
}

func (p *postgresDocumentDB) DeleteDatabase(ctx context.Context, projectID, id string) error {
	internalID, err := p.resolveInternalID(ctx, projectID)
	if err != nil {
		return err
	}
	schema := schemaName(internalID, id)
	if _, err := p.db.DB.ExecContext(ctx, fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE`, quoteIdent(schema))); err != nil {
		return err
	}
	_, err = p.db.NewDelete().Model((*model.DocumentDatabase)(nil)).Where("id = ? AND project_id = ?", id, projectID).Exec(ctx)
	return err
}

func (p *postgresDocumentDB) CreateCollection(ctx context.Context, projectID, databaseID, collectionID, name string, attrs []databases.Attribute, idxs []databases.Index) error {
	internalID, err := p.resolveInternalID(ctx, projectID)
	if err != nil {
		return err
	}
	schema := schemaName(internalID, databaseID)
	if err := p.ensureSchemaAndPerms(ctx, schema); err != nil {
		return err
	}

	if err := p.createCollectionTable(ctx, schema, collectionID, internalID, attrs); err != nil {
		return err
	}
	for _, idx := range idxs {
		if err := p.createCollectionIndex(ctx, schema, collectionID, idx); err != nil {
			return err
		}
	}

	return p.createCollectionMetadata(ctx, projectID, databaseID, collectionID, name, attrs, idxs)
}

func (p *postgresDocumentDB) GetCollection(ctx context.Context, projectID, databaseID, collectionID string) (*databases.Collection, error) {
	m := new(model.DocumentCollection)
	err := p.conn(ctx).NewSelect().Model(m).
		Where("project_id = ? AND database_id = ? AND id = ?", projectID, databaseID, collectionID).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return p.mapCollection(ctx, m)
}

func (p *postgresDocumentDB) ListCollections(ctx context.Context, projectID, databaseID string) ([]databases.Collection, error) {
	var ms []model.DocumentCollection
	err := p.db.NewSelect().Model(&ms).
		Where("project_id = ? AND database_id = ?", projectID, databaseID).
		Order("created_at DESC").Scan(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]databases.Collection, len(ms))
	for i := range ms {
		c, err := p.mapCollection(ctx, &ms[i])
		if err != nil {
			return nil, err
		}
		out[i] = *c
	}
	return out, nil
}

func (p *postgresDocumentDB) DeleteCollection(ctx context.Context, projectID, databaseID, collectionID string) error {
	internalID, err := p.resolveInternalID(ctx, projectID)
	if err != nil {
		return err
	}
	schema := schemaName(internalID, databaseID)
	if _, err := p.db.DB.ExecContext(ctx, fmt.Sprintf(`DROP TABLE IF EXISTS %s CASCADE`, tableName(schema, collectionID))); err != nil {
		return err
	}
	_, err = p.db.NewDelete().Model((*model.DocumentCollection)(nil)).
		Where("project_id = ? AND database_id = ? AND id = ?", projectID, databaseID, collectionID).Exec(ctx)
	return err
}

func (p *postgresDocumentDB) CreateAttribute(ctx context.Context, projectID, databaseID, collectionID string, attr databases.Attribute) error {
	internalID, err := p.resolveInternalID(ctx, projectID)
	if err != nil {
		return err
	}
	schema := schemaName(internalID, databaseID)
	colSQL := attributeColumnSQL(attr)
	if _, err := p.db.DB.ExecContext(ctx, fmt.Sprintf(`ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s`, tableName(schema, collectionID), colSQL)); err != nil {
		return err
	}
	m := &model.DocumentAttribute{
		ID:           attr.ID,
		CollectionID: collectionID,
		DatabaseID:   databaseID,
		ProjectID:    projectID,
		Key:          attr.Key,
		Type:         attr.Type,
		Required:     attr.Required,
		IsArray:      attr.Array,
		CreatedAt:    time.Now(),
	}
	if attr.Size > 0 {
		m.Size = &attr.Size
	}
	_, err = p.db.NewInsert().Model(m).Exec(ctx)
	return err
}

func (p *postgresDocumentDB) CreateIndex(ctx context.Context, projectID, databaseID, collectionID string, idx databases.Index) error {
	internalID, err := p.resolveInternalID(ctx, projectID)
	if err != nil {
		return err
	}
	schema := schemaName(internalID, databaseID)
	if err := p.createCollectionIndex(ctx, schema, collectionID, idx); err != nil {
		return err
	}
	m := &model.DocumentIndex{
		ID:           idx.ID,
		CollectionID: collectionID,
		DatabaseID:   databaseID,
		ProjectID:    projectID,
		Type:         idx.Type,
		Attributes:   idx.Attributes,
		Orders:       idx.Orders,
		CreatedAt:    time.Now(),
	}
	_, err = p.db.NewInsert().Model(m).Exec(ctx)
	return err
}

func (p *postgresDocumentDB) CreateDocument(ctx context.Context, projectID, databaseID, collectionID string, doc databases.Document, perms []databases.Permission) (databases.Document, error) {
	internalID, err := p.resolveInternalID(ctx, projectID)
	if err != nil {
		return doc, err
	}
	if doc.ID == "" {
		doc.ID = idgen.UUID().String()
	}
	schema := schemaName(internalID, databaseID)
	tbl := tableName(schema, collectionID)
	columns, placeholders, args := buildInsertParts(doc)
	args = append([]any{doc.ID}, args...)
	allPlaceholders := "?"
	if columns != "" {
		allPlaceholders = "?, " + placeholders
		columns = ", " + columns
	}
	sql := fmt.Sprintf(`INSERT INTO %s (_id%s) VALUES (%s)`, tbl, columns, allPlaceholders)
	if _, err := p.db.DB.ExecContext(ctx, sql, args...); err != nil {
		return doc, fmt.Errorf("insert document: %w", err)
	}
	if err := p.setPermissions(ctx, schema, collectionID, doc.ID, internalID, perms); err != nil {
		return doc, err
	}
	created, err := p.GetDocument(ctx, projectID, databaseID, collectionID, doc.ID, databases.SystemRoles)
	if err != nil {
		return doc, err
	}
	if created == nil {
		return doc, errors.New("document not found after insert")
	}
	return *created, nil
}

func (p *postgresDocumentDB) GetDocument(ctx context.Context, projectID, databaseID, collectionID, docID string, roles []string) (*databases.Document, error) {
	if err := validateDocID(docID); err != nil {
		return nil, err
	}
	internalID, err := p.resolveInternalID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	schema := schemaName(internalID, databaseID)
	row := p.db.DB.QueryRowContext(ctx, fmt.Sprintf(`SELECT to_jsonb(d.*) AS doc FROM %s d WHERE d._id = ? AND d._tenant = ?`, tableName(schema, collectionID)), docID, internalID)
	doc, err := scanDocumentJSON(row)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}
	if err := p.checkDocumentPermission(ctx, schema, collectionID, docID, internalID, "read", roles); err != nil {
		return nil, err
	}
	return doc, nil
}

func (p *postgresDocumentDB) UpdateDocument(ctx context.Context, projectID, databaseID, collectionID string, doc databases.Document, perms []databases.Permission, roles []string) (databases.Document, error) {
	if err := validateDocID(doc.ID); err != nil {
		return doc, err
	}
	internalID, err := p.resolveInternalID(ctx, projectID)
	if err != nil {
		return doc, err
	}
	schema := schemaName(internalID, databaseID)
	if err := p.checkDocumentPermission(ctx, schema, collectionID, doc.ID, internalID, "update", roles); err != nil {
		return doc, err
	}
	tbl := tableName(schema, collectionID)
	setParts, args := buildUpdateParts(doc)
	if len(setParts) == 0 {
		return doc, fmt.Errorf("no fields to update")
	}
	args = append(args, doc.ID, internalID)
	sql := fmt.Sprintf(`UPDATE %s SET %s WHERE _id = ? AND _tenant = ?`, tbl, strings.Join(setParts, ", "))
	if _, err := p.db.DB.ExecContext(ctx, sql, args...); err != nil {
		return doc, fmt.Errorf("update document: %w", err)
	}
	if len(perms) > 0 {
		if err := p.clearPermissions(ctx, schema, collectionID, doc.ID, internalID); err != nil {
			return doc, err
		}
		if err := p.setPermissions(ctx, schema, collectionID, doc.ID, internalID, perms); err != nil {
			return doc, err
		}
	}
	updated, err := p.GetDocument(ctx, projectID, databaseID, collectionID, doc.ID, roles)
	if err != nil {
		return doc, err
	}
	if updated == nil {
		return doc, errors.New("document not found after update")
	}
	return *updated, nil
}

func (p *postgresDocumentDB) DeleteDocument(ctx context.Context, projectID, databaseID, collectionID, docID string, roles []string) error {
	if err := validateDocID(docID); err != nil {
		return err
	}
	internalID, err := p.resolveInternalID(ctx, projectID)
	if err != nil {
		return err
	}
	schema := schemaName(internalID, databaseID)
	if err := p.checkDocumentPermission(ctx, schema, collectionID, docID, internalID, "delete", roles); err != nil {
		return err
	}
	if err := p.clearPermissions(ctx, schema, collectionID, docID, internalID); err != nil {
		return err
	}
	_, err = p.db.DB.ExecContext(ctx, fmt.Sprintf(`DELETE FROM %s WHERE _id = ? AND _tenant = ?`, tableName(schema, collectionID)), docID, internalID)
	return err
}

func (p *postgresDocumentDB) ListDocuments(ctx context.Context, projectID, databaseID, collectionID string, q databases.Query, roles []string) (*databases.DocumentList, error) {
	internalID, err := p.resolveInternalID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	parsed, err := query.ParseMany(q.Queries)
	if err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	schema := schemaName(internalID, databaseID)
	tbl := tableName(schema, collectionID)
	permsTable := permsTableName(schema)

	whereParts := []string{"d._tenant = ?"}
	args := []any{internalID}
	if !bypassDocumentPermissions(roles) {
		whereParts = append(whereParts, fmt.Sprintf(`EXISTS (SELECT 1 FROM %s p WHERE p._collection = ? AND p._document = d._id AND p._type = 'read' AND p._permission = ANY(?::text[]))`, permsTable))
		args = append(args, collectionID, pgTextArray(expandPermissionRoles(roles)))
	}

	filterWhere, filterArgs, orderSQL, err := buildAppwriteQuery(parsed)
	if err != nil {
		return nil, err
	}
	if filterWhere != "" {
		whereParts = append(whereParts, filterWhere)
		args = append(args, filterArgs...)
	}

	limit := parsed.Limit
	if limit == 0 {
		limit = int(q.PageSize)
	}
	if limit == 0 {
		limit = 50
	}
	if limit > maxQueryLimit {
		limit = maxQueryLimit
	}
	offset := parsed.Offset
	if q.PageToken != "" {
		if off, err := crud.DecodePageToken(q.PageToken); err == nil {
			offset = int(off)
		}
	}

	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM %s d WHERE %s`, tbl, strings.Join(whereParts, " AND "))
	var total int64
	if err := p.db.DB.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, err
	}

	querySQL := fmt.Sprintf(`SELECT to_jsonb(d.*) AS doc FROM %s d WHERE %s %s LIMIT ? OFFSET ?`, tbl, strings.Join(whereParts, " AND "), orderSQL)
	args = append(args, limit, offset)

	rows, err := p.db.DB.QueryContext(ctx, querySQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []databases.Document
	for rows.Next() {
		doc, err := scanDocumentJSON(rows)
		if err != nil {
			return nil, err
		}
		docs = append(docs, *doc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	next := ""
	if len(docs) > 0 && int64(offset+len(docs)) < total {
		next = crud.EncodePageToken(offset + len(docs))
	}
	return &databases.DocumentList{
		Documents:     docs,
		TotalCount:    total,
		NextPageToken: next,
	}, nil
}

func (p *postgresDocumentDB) CountDocuments(ctx context.Context, projectID, databaseID, collectionID string, queries []string, roles []string) (int64, error) {
	internalID, err := p.resolveInternalID(ctx, projectID)
	if err != nil {
		return 0, err
	}
	parsed, err := query.ParseMany(queries)
	if err != nil {
		return 0, fmt.Errorf("invalid query: %w", err)
	}
	schema := schemaName(internalID, databaseID)
	tbl := tableName(schema, collectionID)
	permsTable := permsTableName(schema)

	whereParts := []string{"d._tenant = ?"}
	args := []any{internalID}
	if !bypassDocumentPermissions(roles) {
		whereParts = append(whereParts, fmt.Sprintf(`EXISTS (SELECT 1 FROM %s p WHERE p._collection = ? AND p._document = d._id AND p._type = 'read' AND p._permission = ANY(?::text[]))`, permsTable))
		args = append(args, collectionID, pgTextArray(expandPermissionRoles(roles)))
	}
	filterWhere, filterArgs, _, err := buildAppwriteQuery(parsed)
	if err != nil {
		return 0, err
	}
	if filterWhere != "" {
		whereParts = append(whereParts, filterWhere)
		args = append(args, filterArgs...)
	}

	var total int64
	sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s d WHERE %s`, tbl, strings.Join(whereParts, " AND "))
	err = p.db.DB.QueryRowContext(ctx, sql, args...).Scan(&total)
	return total, err
}

func (p *postgresDocumentDB) EnsureSystemCollections(ctx context.Context, projectID string, internalID int64) error {
	dbID := "default"
	schema := schemaName(internalID, dbID)
	if err := p.ensureSchemaAndPerms(ctx, schema); err != nil {
		return err
	}

	// Ensure default database metadata row exists.
	exists, err := p.conn(ctx).NewSelect().Model((*model.DocumentDatabase)(nil)).
		Where("id = ? AND project_id = ?", dbID, projectID).Exists(ctx)
	if err != nil {
		return err
	}
	if !exists {
		m := &model.DocumentDatabase{ID: dbID, ProjectID: projectID, Name: "default", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		if _, err := p.conn(ctx).NewInsert().Model(m).Exec(ctx); err != nil {
			return err
		}
	}

	for _, spec := range systemCollectionSpecs(projectID) {
		coll, err := p.GetCollection(ctx, projectID, dbID, spec.id)
		if err != nil {
			return err
		}
		if coll != nil {
			continue
		}
		if err := p.CreateCollection(ctx, projectID, dbID, spec.id, spec.name, spec.attrs, spec.indexes); err != nil {
			return fmt.Errorf("create system collection %s: %w", spec.id, err)
		}
		if err := p.setCollectionPermissions(ctx, projectID, dbID, spec.id, spec.permissions); err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func schemaName(internalID int64, databaseID string) string {
	return fmt.Sprintf("fleet_%d_%s", internalID, databaseID)
}

func tableName(schema, collectionID string) string {
	return quoteIdent(schema) + "." + quoteIdent(collectionID)
}

func permsTableName(schema string) string {
	return quoteIdent(schema) + "." + quoteIdent("_perms")
}

func pgTextArray(items []string) string {
	var parts []string
	for _, item := range items {
		parts = append(parts, `"`+strings.ReplaceAll(item, `"`, `""`)+`"`)
	}
	return `{` + strings.Join(parts, ",") + `}`
}

func (p *postgresDocumentDB) resolveInternalID(ctx context.Context, projectID string) (int64, error) {
	var internalID int64
	err := p.conn(ctx).NewSelect().Model((*model.Project)(nil)).Column("internal_id").Where("id = ?", projectID).Scan(ctx, &internalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("project not found: %s", projectID)
		}
		return 0, err
	}
	return internalID, nil
}

func (p *postgresDocumentDB) ensureSchemaAndPerms(ctx context.Context, schema string) error {
	if _, err := p.conn(ctx).ExecContext(ctx, fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, quoteIdent(schema))); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}
	return p.ensurePermsTable(ctx, schema)
}

func (p *postgresDocumentDB) ensurePermsTable(ctx context.Context, schema string) error {
	sql := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		_id BIGSERIAL PRIMARY KEY,
		_tenant BIGINT NOT NULL,
		_collection TEXT NOT NULL,
		_document TEXT NOT NULL,
		_type TEXT NOT NULL,
		_permission TEXT NOT NULL,
		_created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE (_tenant, _collection, _document, _type, _permission)
	)`, permsTableName(schema))
	if _, err := p.conn(ctx).ExecContext(ctx, sql); err != nil {
		return fmt.Errorf("create perms table: %w", err)
	}
	idx := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_perms_lookup ON %s (_tenant, _collection, _document, _type)`, permsTableName(schema))
	if _, err := p.conn(ctx).ExecContext(ctx, idx); err != nil {
		return err
	}
	idx2 := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_perms_role ON %s (_tenant, _collection, _type, _permission)`, permsTableName(schema))
	_, err := p.conn(ctx).ExecContext(ctx, idx2)
	return err
}

func (p *postgresDocumentDB) createCollectionTable(ctx context.Context, schema, collectionID string, tenant int64, attrs []databases.Attribute) error {
	cols := []string{
		"_id TEXT PRIMARY KEY",
		fmt.Sprintf("_tenant BIGINT NOT NULL DEFAULT %d", tenant),
		"_created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()",
		"_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()",
		"_created_by TEXT",
		"_updated_by TEXT",
	}
	for _, attr := range attrs {
		cols = append(cols, attributeColumnSQL(attr))
	}
	cols = append(cols, "UNIQUE (_id, _tenant)")
	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n\t\t%s\n\t)", tableName(schema, collectionID), strings.Join(cols, ",\n\t\t"))
	_, err := p.conn(ctx).ExecContext(ctx, sql)
	return err
}

func (p *postgresDocumentDB) createCollectionIndex(ctx context.Context, schema, collectionID string, idx databases.Index) error {
	var cols []string
	for i, attr := range idx.Attributes {
		if !safeNameRe.MatchString(attr) {
			return fmt.Errorf("invalid index attribute: %s", attr)
		}
		order := ""
		if i < len(idx.Orders) && strings.EqualFold(idx.Orders[i], "desc") {
			order = " DESC"
		}
		cols = append(cols, quoteIdent(attr)+order)
	}
	idxName := quoteIdent(fmt.Sprintf("idx_%s_%s", collectionID, idx.ID))
	var sql string
	switch strings.ToLower(idx.Type) {
	case "unique":
		sql = fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s (%s)`, idxName, tableName(schema, collectionID), strings.Join(cols, ", "))
	case "fulltext":
		sql = fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s USING gin(to_tsvector('simple', %s))`, idxName, tableName(schema, collectionID), strings.Join(cols, " || ' ' || "))
	default:
		sql = fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s (%s)`, idxName, tableName(schema, collectionID), strings.Join(cols, ", "))
	}
	_, err := p.conn(ctx).ExecContext(ctx, sql)
	return err
}

func attributeColumnSQL(attr databases.Attribute) string {
	name := quoteIdent(attr.Key)
	if !safeNameRe.MatchString(attr.Key) {
		panic(fmt.Sprintf("invalid attribute key: %s", attr.Key))
	}
	dataType := pgTypeFor(attr.Type, attr.Size)
	parts := []string{name, dataType}
	if attr.Required {
		parts = append(parts, "NOT NULL")
	}
	if attr.Default != nil && !attr.Array {
		parts = append(parts, fmt.Sprintf("DEFAULT %s", formatDefault(attr.Default, attr.Type)))
	}
	return strings.Join(parts, " ")
}

func pgTypeFor(t string, size int) string {
	switch strings.ToLower(t) {
	case "string", "email", "url":
		if size > 0 && size <= 64000 {
			return fmt.Sprintf("VARCHAR(%d)", size)
		}
		return "TEXT"
	case "integer":
		return "BIGINT"
	case "float":
		return "DOUBLE PRECISION"
	case "boolean":
		return "BOOLEAN"
	case "datetime":
		return "TIMESTAMPTZ"
	case "json":
		return "JSONB"
	default:
		return "TEXT"
	}
}

func formatDefault(v any, t string) string {
	switch strings.ToLower(t) {
	case "boolean":
		if b, _ := strconv.ParseBool(fmt.Sprint(v)); b {
			return "TRUE"
		}
		return "FALSE"
	case "integer":
		n, err := strconv.ParseInt(fmt.Sprint(v), 10, 64)
		if err != nil {
			return "0"
		}
		return strconv.FormatInt(n, 10)
	case "float":
		f, err := strconv.ParseFloat(fmt.Sprint(v), 64)
		if err != nil {
			return "0"
		}
		return strconv.FormatFloat(f, 'f', -1, 64)
	default:
		return quoteLiteral(fmt.Sprint(v))
	}
}

func quoteLiteral(s string) string {
	return `'` + strings.ReplaceAll(s, `'`, `''`) + `'`
}

func (p *postgresDocumentDB) createCollectionMetadata(ctx context.Context, projectID, databaseID, collectionID, name string, attrs []databases.Attribute, idxs []databases.Index) error {
	coll := &model.DocumentCollection{
		ID:               collectionID,
		DatabaseID:       databaseID,
		ProjectID:        projectID,
		Name:             name,
		DocumentSecurity: true,
		Permissions:      []string{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if _, err := p.conn(ctx).NewInsert().Model(coll).Exec(ctx); err != nil {
		return err
	}
	for _, attr := range attrs {
		m := &model.DocumentAttribute{
			ID:           attr.ID,
			CollectionID: collectionID,
			DatabaseID:   databaseID,
			ProjectID:    projectID,
			Key:          attr.Key,
			Type:         attr.Type,
			Required:     attr.Required,
			IsArray:      attr.Array,
			CreatedAt:    time.Now(),
		}
		if attr.Size > 0 {
			m.Size = &attr.Size
		}
		if _, err := p.conn(ctx).NewInsert().Model(m).Exec(ctx); err != nil {
			return err
		}
	}
	for _, idx := range idxs {
		m := &model.DocumentIndex{
			ID:           idx.ID,
			CollectionID: collectionID,
			DatabaseID:   databaseID,
			ProjectID:    projectID,
			Type:         idx.Type,
			Attributes:   idx.Attributes,
			Orders:       idx.Orders,
			CreatedAt:    time.Now(),
		}
		if _, err := p.conn(ctx).NewInsert().Model(m).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (p *postgresDocumentDB) mapCollection(ctx context.Context, m *model.DocumentCollection) (*databases.Collection, error) {
	var attrs []model.DocumentAttribute
	if err := p.conn(ctx).NewSelect().Model(&attrs).
		Where("project_id = ? AND database_id = ? AND collection_id = ?", m.ProjectID, m.DatabaseID, m.ID).
		Scan(ctx); err != nil {
		return nil, err
	}
	var idxs []model.DocumentIndex
	if err := p.conn(ctx).NewSelect().Model(&idxs).
		Where("project_id = ? AND database_id = ? AND collection_id = ?", m.ProjectID, m.DatabaseID, m.ID).
		Scan(ctx); err != nil {
		return nil, err
	}
	c := &databases.Collection{
		ID:               m.ID,
		DatabaseID:       m.DatabaseID,
		ProjectID:        m.ProjectID,
		Name:             m.Name,
		DocumentSecurity: m.DocumentSecurity,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
	for _, p := range m.Permissions {
		c.Permissions = append(c.Permissions, parsePermission(p))
	}
	for _, a := range attrs {
		attr := databases.Attribute{ID: a.ID, Key: a.Key, Type: a.Type, Required: a.Required, Array: a.IsArray}
		if a.Size != nil {
			attr.Size = *a.Size
		}
		c.Attributes = append(c.Attributes, attr)
	}
	for _, i := range idxs {
		c.Indexes = append(c.Indexes, databases.Index{ID: i.ID, Type: i.Type, Attributes: i.Attributes, Orders: i.Orders})
	}
	return c, nil
}

func parsePermission(s string) databases.Permission {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return databases.Permission{}
	}
	return databases.Permission{Type: parts[0], Role: parts[1]}
}

func (p *postgresDocumentDB) setCollectionPermissions(ctx context.Context, projectID, databaseID, collectionID string, perms []databases.Permission) error {
	var raw []string
	for _, perm := range perms {
		raw = append(raw, fmt.Sprintf("%s:%s", perm.Type, perm.Role))
	}
	_, err := p.conn(ctx).ExecContext(ctx,
		`UPDATE document_collections SET permissions = ? WHERE project_id = ? AND database_id = ? AND id = ?`,
		pgdialect.Array(raw), projectID, databaseID, collectionID)
	return err
}

func (p *postgresDocumentDB) setPermissions(ctx context.Context, schema, collectionID, documentID string, tenant int64, perms []databases.Permission) error {
	if len(perms) == 0 {
		return nil
	}
	base := fmt.Sprintf(`INSERT INTO %s (_tenant, _collection, _document, _type, _permission) VALUES `, permsTableName(schema))
	var vals []string
	var args []any
	for range perms {
		vals = append(vals, "(?, ?, ?, ?, ?)")
	}
	for _, perm := range perms {
		args = append(args, tenant, collectionID, documentID, perm.Type, perm.Role)
	}
	sql := base + strings.Join(vals, ", ") + " ON CONFLICT DO NOTHING"
	_, err := p.db.DB.ExecContext(ctx, sql, args...)
	return err
}

func (p *postgresDocumentDB) clearPermissions(ctx context.Context, schema, collectionID, documentID string, tenant int64) error {
	_, err := p.db.DB.ExecContext(ctx, fmt.Sprintf(`DELETE FROM %s WHERE _tenant = ? AND _collection = ? AND _document = ?`, permsTableName(schema)), tenant, collectionID, documentID)
	return err
}

func buildInsertParts(doc databases.Document) (columns string, placeholders string, args []any) {
	if len(doc.Data) == 0 {
		return "", "", nil
	}
	var cols []string
	var phs []string
	for k, v := range doc.Data {
		if !safeNameRe.MatchString(k) {
			continue
		}
		cols = append(cols, quoteIdent(k))
		phs = append(phs, "?")
		args = append(args, v)
	}
	return strings.Join(cols, ", "), strings.Join(phs, ", "), args
}

func buildUpdateParts(doc databases.Document) (setParts []string, args []any) {
	for k, v := range doc.Data {
		if !safeNameRe.MatchString(k) || strings.HasPrefix(k, "_") {
			continue
		}
		setParts = append(setParts, fmt.Sprintf("%s = ?", quoteIdent(k)))
		args = append(args, v)
	}
	if len(setParts) == 0 {
		return nil, nil
	}
	setParts = append(setParts, "_updated_at = ?")
	args = append(args, time.Now())
	return setParts, args
}

func scanDocumentJSON(scanner interface{ Scan(dest ...any) error }) (*databases.Document, error) {
	var raw []byte
	if err := scanner.Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	doc := &databases.Document{Data: make(map[string]any)}
	if v, ok := payload["_id"].(string); ok {
		doc.ID = v
	}
	if v, ok := payload["_tenant"].(float64); ok {
		doc.Tenant = int64(v)
	}
	if v, ok := payload["_created_at"].(string); ok {
		doc.CreatedAt, _ = time.Parse(time.RFC3339Nano, v)
	}
	if v, ok := payload["_updated_at"].(string); ok {
		doc.UpdatedAt, _ = time.Parse(time.RFC3339Nano, v)
	}
	if v, ok := payload["_created_by"].(string); ok {
		doc.CreatedBy = v
	}
	if v, ok := payload["_updated_by"].(string); ok {
		doc.UpdatedBy = v
	}
	for k, v := range payload {
		if strings.HasPrefix(k, "_") {
			continue
		}
		doc.Data[k] = v
	}
	return doc, nil
}

func bypassDocumentPermissions(roles []string) bool {
	for _, r := range roles {
		switch r {
		case "__system__", "keys", "owner", "admin":
			return true
		}
	}
	return false
}

func expandPermissionRoles(roles []string) []string {
	seen := make(map[string]struct{}, len(roles)+2)
	out := make([]string, 0, len(roles)+2)
	for _, r := range roles {
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		out = append(out, r)
	}
	for _, r := range []string{"any", "users"} {
		if _, ok := seen[r]; !ok {
			out = append(out, r)
		}
	}
	return out
}

func (p *postgresDocumentDB) checkDocumentPermission(ctx context.Context, schema, collectionID, docID string, tenant int64, permType string, roles []string) error {
	if bypassDocumentPermissions(roles) {
		return nil
	}
	permsTable := permsTableName(schema)
	sql := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s p WHERE p._tenant = ? AND p._collection = ? AND p._document = ? AND p._type = ? AND p._permission = ANY(?::text[]))`, permsTable)
	var ok bool
	if err := p.db.DB.QueryRowContext(ctx, sql, tenant, collectionID, docID, permType, pgTextArray(expandPermissionRoles(roles))).Scan(&ok); err != nil {
		return err
	}
	if !ok {
		return ErrPermissionDenied
	}
	return nil
}

func validateDocID(docID string) error {
	if docID == "" {
		return errors.New("document id is required")
	}
	if !docIDRe.MatchString(docID) {
		return fmt.Errorf("invalid document id: %s", docID)
	}
	return nil
}

func hasAdminRole(roles []string) bool {
	return bypassDocumentPermissions(roles)
}

func mapQueryField(field string) string {
	switch field {
	case "$id", "_id":
		return "_id"
	case "$createdAt", "_created_at":
		return "_created_at"
	case "$updatedAt", "_updated_at":
		return "_updated_at"
	}
	return field
}

func buildAppwriteQuery(parsed *query.Query) (string, []any, string, error) {
	var conds []string
	var args []any
	for _, f := range parsed.Filters {
		field := mapQueryField(f.Attribute)
		if !safeNameRe.MatchString(field) {
			return "", nil, "", fmt.Errorf("invalid query field: %s", f.Attribute)
		}
		col := "d." + quoteIdent(field)
		switch f.Op {
		case "equal":
			if len(f.Values) == 1 {
				conds = append(conds, fmt.Sprintf("%s = ?", col))
				args = append(args, f.Values[0])
			} else {
				conds = append(conds, fmt.Sprintf("%s = ANY(?::text[])", col))
				args = append(args, pgTextArray(f.Values))
			}
		case "notEqual":
			if len(f.Values) == 1 {
				conds = append(conds, fmt.Sprintf("%s != ?", col))
				args = append(args, f.Values[0])
			} else {
				conds = append(conds, fmt.Sprintf("%s != ALL(?::text[])", col))
				args = append(args, pgTextArray(f.Values))
			}
		case "lessThan":
			conds = append(conds, fmt.Sprintf("%s < ?", col))
			args = append(args, f.Values[0])
		case "lessThanEqual":
			conds = append(conds, fmt.Sprintf("%s <= ?", col))
			args = append(args, f.Values[0])
		case "greaterThan":
			conds = append(conds, fmt.Sprintf("%s > ?", col))
			args = append(args, f.Values[0])
		case "greaterThanEqual":
			conds = append(conds, fmt.Sprintf("%s >= ?", col))
			args = append(args, f.Values[0])
		case "contains":
			conds = append(conds, fmt.Sprintf("%s ILIKE ?", col))
			args = append(args, "%"+f.Values[0]+"%")
		case "startsWith":
			conds = append(conds, fmt.Sprintf("%s ILIKE ?", col))
			args = append(args, f.Values[0]+"%")
		case "endsWith":
			conds = append(conds, fmt.Sprintf("%s ILIKE ?", col))
			args = append(args, "%"+f.Values[0])
		case "search":
			conds = append(conds, fmt.Sprintf("to_tsvector('simple', %s::text) @@ plainto_tsquery('simple', ?)", col))
			args = append(args, f.Values[0])
		case "isNull":
			conds = append(conds, fmt.Sprintf("%s IS NULL", col))
		case "isNotNull":
			conds = append(conds, fmt.Sprintf("%s IS NOT NULL", col))
		case "between":
			if len(f.Values) != 2 {
				return "", nil, "", fmt.Errorf("between requires 2 values")
			}
			conds = append(conds, fmt.Sprintf("%s BETWEEN ? AND ?", col))
			args = append(args, f.Values[0], f.Values[1])
		default:
			return "", nil, "", fmt.Errorf("unsupported filter operator: %s", f.Op)
		}
	}

	orderSQL := "ORDER BY d._created_at DESC"
	if len(parsed.Orders) > 0 {
		var parts []string
		for _, o := range parsed.Orders {
			field := mapQueryField(o.Attribute)
			if !safeNameRe.MatchString(field) {
				continue
			}
			dir := "ASC"
			if o.Desc {
				dir = "DESC"
			}
			parts = append(parts, fmt.Sprintf("d.%s %s", quoteIdent(field), dir))
		}
		if len(parts) > 0 {
			orderSQL = "ORDER BY " + strings.Join(parts, ", ") + ", d._created_at DESC"
		}
	}

	where := ""
	if len(conds) > 0 {
		where = "(" + strings.Join(conds, " AND ") + ")"
	}
	return where, args, orderSQL, nil
}
