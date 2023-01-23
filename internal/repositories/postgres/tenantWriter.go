package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Masterminds/squirrel"
	otelCodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/protobuf/types/known/timestamppb"

	db "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TenantWriter - Structure for Tenant Writer
type TenantWriter struct {
	database *db.Postgres
	// options
	txOptions sql.TxOptions
	// logger
	logger logger.Interface
}

// NewTenantWriter - Creates a new TenantWriter
func NewTenantWriter(database *db.Postgres, logger logger.Interface) *TenantWriter {
	return &TenantWriter{
		database:  database,
		txOptions: sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false},
		logger:    logger,
	}
}

// CreateTenant - Creates a new Tenant
func (w *TenantWriter) CreateTenant(ctx context.Context, name string) (result *base.Tenant, err error) {
	ctx, span := tracer.Start(ctx, "tenant-writer.create-tenant")
	defer span.End()

	var id uint64
	var createdAt time.Time

	query := w.database.Builder.Insert(TenantsTable).Columns("name").Values(name).Suffix("RETURNING id, created_at").RunWith(w.database.DB)

	err = query.QueryRowContext(ctx).Scan(&id, &createdAt)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	return &base.Tenant{
		Id:        id,
		Name:      name,
		CreatedAt: timestamppb.New(createdAt),
	}, nil
}

// DeleteTenant - Deletes a Tenant
func (w *TenantWriter) DeleteTenant(ctx context.Context, tenantID uint64) (result *base.Tenant, err error) {
	ctx, span := tracer.Start(ctx, "tenant-writer.delete-tenant")
	defer span.End()

	var name string
	var createdAt time.Time

	query := w.database.Builder.Delete(TenantsTable).Where(squirrel.Eq{"id": tenantID}).Suffix("RETURNING name, created_at").RunWith(w.database.DB)
	err = query.QueryRowContext(ctx).Scan(&name, &createdAt)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	return &base.Tenant{
		Id:        tenantID,
		Name:      name,
		CreatedAt: timestamppb.New(createdAt),
	}, nil
}