package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kolakdd/cache_storage/golang/apiError"
	"github.com/kolakdd/cache_storage/golang/models"
	uuid "github.com/satori/go.uuid"
)

type ObjRepo interface {
	CreateTX() *sql.Tx
	Create(tx *sql.Tx, dto *models.UploadObjectDtoMeta, ownerID uuid.UUID, size int64) (*models.Object, *apiError.BackendErrorInternal)
}

type objRepo struct {
	db *sql.DB
}

func NewObjRepo(db *sql.DB) ObjRepo {
	return &objRepo{db}
}

func (r *objRepo) Create(tx *sql.Tx, dto *models.UploadObjectDtoMeta, ownerID uuid.UUID, size int64) (*models.Object, *apiError.BackendErrorInternal) {
	objDB := models.NewObjectDB(ownerID, dto.Name, dto.Mime, dto.Public, size)

	q := `
		INSERT INTO "Object" 
		(id, owner_id, name, mimetype, size, upload_s3, is_deleted, eliminated)
		VALUES 
		($1, $2, $3, $4, $5 ,$6 ,$7, $8)
		`
	_, execErr := tx.Exec(q, objDB.ID, objDB.OwnerID, objDB.Name, objDB.Mimetype, objDB.Size, objDB.UploadS3, objDB.IsDeleted, objDB.Eliminated)
	if execErr != nil {
		fmt.Println("failed to execute query: %w", execErr)
		return nil, apiError.InternalError
	}
	return &objDB, nil
}

func (r *objRepo) CreateTX() *sql.Tx {
	ctx := context.Background()
	tx, _ := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault})
	return tx
}
