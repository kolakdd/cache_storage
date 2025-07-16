package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/kolakdd/cache_storage/golang/apiError"
	"github.com/kolakdd/cache_storage/golang/models"
	uuid "github.com/satori/go.uuid"
)

type ObjRepo interface {
	CreateTX() *sql.Tx
	Create(tx *sql.Tx, dto *models.UploadObjectDtoMeta, ownerID uuid.UUID, size int64) (*models.Object, *apiError.BackendErrorInternal)
	GetList(dto *models.GetListObjectsDto) ([]models.Object, *apiError.BackendErrorInternal)
	CheckAccess(objID uuid.UUID, userID uuid.UUID) (uuid.UUID, string, *apiError.BackendErrorInternal)
	DeleteWithAccess(objID uuid.UUID) *apiError.BackendErrorInternal
}

type objRepo struct {
	db *sql.DB
}

func NewObjRepo(db *sql.DB) ObjRepo {
	return &objRepo{db}
}

// Создание объекта
func (r *objRepo) Create(tx *sql.Tx, dto *models.UploadObjectDtoMeta, ownerID uuid.UUID, size int64) (*models.Object, *apiError.BackendErrorInternal) {
	objDB := models.NewObjectDB(ownerID, dto.Name, dto.Mime, dto.Public, size)
	q := `
		INSERT INTO "Object" 
		(id, owner_id, name, mimetype, size, upload_s3, is_deleted, eliminated, public)
		VALUES 
		($1, $2, $3, $4, $5 ,$6 ,$7, $8, $9)
		`
	_, execErr := tx.Exec(q, objDB.ID, objDB.OwnerID, objDB.Name, objDB.Mimetype, objDB.Size, objDB.UploadS3, objDB.IsDeleted, objDB.Eliminated, dto.Public)
	if execErr != nil {
		fmt.Println("failed to execute query: %w", execErr)
		return nil, apiError.InternalError
	}

	return &objDB, nil
}

// Проверка наличия доступа у пользователя по id.
func (r *objRepo) CheckAccess(objID uuid.UUID, userID uuid.UUID) (uuid.UUID, string, *apiError.BackendErrorInternal) {
	q := `
		SELECT "Object".owner_id, "Object".name FROM "Object"
		JOIN "UserXObject" ON "Object".id = "UserXObject".object_id
		WHERE "UserXObject".user_id = $1 AND "UserXObject".object_id = $2
		`
	var ownerID uuid.UUID // владелец файла
	var fileName string   // имя файла

	row := r.db.QueryRow(q, userID, objID)
	switch err := row.Scan(&ownerID, &fileName); err {
	case sql.ErrNoRows:
		return ownerID, "", apiError.NotFound
	case nil:
		return ownerID, fileName, nil
	default:
		fmt.Println(err)
		return ownerID, "", apiError.InternalError
	}
}

// Удаление доступов к удаленному файлу и изменение флага eliminated на true
func (r *objRepo) DeleteWithAccess(objID uuid.UUID) *apiError.BackendErrorInternal {
	tx := r.CreateTX()
	q :=
		`
		DELETE FROM "UserXObject"
		WHERE "UserXObject".object_id = $1
		`
	_, err := tx.Exec(q, objID)
	if err != nil {
		return apiError.InternalError
	}
	q =
		`
		UPDATE "Object"
		SET eliminated = true
		WHERE id = $1
		`
	_, err = tx.Exec(q, objID)
	if err != nil {
		return apiError.InternalError
	}
	if errTx := tx.Commit(); errTx != nil {
		fmt.Println("failed to commit transaction: %w", errTx)
		return apiError.InternalError
	}
	return nil
}

// Получение списка объектов
func (r *objRepo) GetList(dto *models.GetListObjectsDto) ([]models.Object, *apiError.BackendErrorInternal) {
	var q string
	var rows *sql.Rows
	var err error
	var binds []any
	q = `
		SELECT "Object".* 
		FROM "Object"
		JOIN "UserXObject" ON "Object".id = "UserXObject".object_id
	`
	if dto.Login == "" {
		q += `WHERE ("UserXObject".user_id = $1`
		binds = append(binds, dto.UserID)
	} else {
		q += `JOIN "User" ON "UserXObject".user_id = "User".id WHERE ("User".login = $1`
		binds = append(binds, dto.Login)
	}
	binds = append(binds, dto.Limit, dto.Offset)
	q += ` OR "Object".public = true) AND ("Object".eliminated = false`
	if dto.Key != "" {
		q += ` AND "Object".` + dto.Key + ` = $4`
		binds = append(binds, dto.Value)
	}
	q += ` )`
	q += ` ORDER BY "Object".name, "Object".created_at `
	q += ` LIMIT $2 OFFSET $3`
	fmt.Println(q)
	fmt.Println(binds...)
	rows, err = r.db.Query(q, binds...)

	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return nil, apiError.InvalidObjectKey
		}
		return nil, apiError.InternalError
	}
	defer rows.Close()
	objList := []models.Object{}
	for rows.Next() {
		var o models.Object
		if err := rows.Scan(&o.ID, &o.OwnerID, &o.Name, &o.Mimetype, &o.Public, &o.Size, &o.UploadS3, &o.IsDeleted, &o.Eliminated, &o.CreatedAt, &o.UpdatedAt); err != nil {
			fmt.Println("marchal err:", rows)

			fmt.Println("marchal err:", err)
			return nil, apiError.InternalError
		}
		objList = append(objList, o)
	}
	return objList, nil
}

func (r *objRepo) CreateTX() *sql.Tx {
	ctx := context.Background()
	tx, _ := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault})
	return tx
}
