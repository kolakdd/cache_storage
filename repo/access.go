package repo

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/kolakdd/cache_storage/apiError"
	uuid "github.com/satori/go.uuid"
)

// AccessRepo отвечает за взаимодействие с доступами к объекту
type AccessRepo interface {
	CreateMany(tx *sql.Tx, objID uuid.UUID, grant []string) *apiError.BackendErrorInternal
}

type accessRepo struct {
	db *sql.DB
}

func NewAccessRepo(db *sql.DB) AccessRepo {
	return &accessRepo{db}
}
func (r *accessRepo) CreateMany(tx *sql.Tx, objID uuid.UUID, grant []string) *apiError.BackendErrorInternal {
	if len(grant) == 0 {
		return nil
	}

	args := make([]string, len(grant))
	params := make([]interface{}, len(grant))
	for i, login := range grant {
		args[i] = fmt.Sprintf("$%d", i+1)
		params[i] = login
	}
	// по хорошему принимать не логин, а сразу id
	q := `SELECT id FROM "User" WHERE login IN (` + strings.Join(args, ", ") + `)`
	rows, err := r.db.Query(q, params...)
	if err != nil {
		return apiError.InternalError
	}
	defer rows.Close()

	userIDS := []interface{}{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return apiError.InternalError
		}
		userIDS = append(userIDS, id)
	}

	if len(userIDS) == 0 {
		return nil
	}

	// по полученным id создаем доступы
	values := make([]string, len(userIDS))
	insertParams := make([]interface{}, len(userIDS)+1)
	insertParams[0] = objID
	for i := 0; i < len(userIDS); i++ {
		values[i] = fmt.Sprintf("($%d, $1)", i+2)
		insertParams[i+1] = userIDS[i]
	}

	q = `INSERT INTO "UserXObject" (user_id, object_id) VALUES ` + strings.Join(values, ",")
	_, execErr := tx.Exec(q, insertParams...)
	if execErr != nil {
		if strings.Contains(execErr.Error(), "duplicate key value violates unique constraint") {
			return apiError.UserAlreadyExist
		}
		return apiError.InternalError
	}
	return nil
}
