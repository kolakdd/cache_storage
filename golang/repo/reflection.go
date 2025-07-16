package repo

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/kolakdd/cache_storage/golang/apiError"
	"github.com/kolakdd/cache_storage/golang/models"
)

// CRUDCreate предоставляет возможность создавать модель в базе
// Модель models.Cruded должна иметь алиасы тега db:"" и иметь название как в самой БД
// пример models.User:
//
//	User struct {
//		ID           uuid.UUID  `json:"id" db:"id"`
//		Login        string     `json:"login" db:"login"`
//		HashPassword string     `json:"hashPassword" db:"hash_password"`
//		CreatedAt    *time.Time `json:"createdAt" db:"created_at"`
//		UpdatedAt    *time.Time `json:"updatedAt" db:"updated_at"`
//	}
func CRUDCreate[T models.Cruded](db *sql.DB, model T) *apiError.BackendErrorInternal {
	v := reflect.ValueOf(&model).Elem()
	t := reflect.TypeOf(&model).Elem()

	len := t.NumField()
	fields := make([]string, 0, len)
	args := make([]string, 0, len)
	values := make([]any, 0, len)

	n := 1
	for i := range len {
		field := v.Field(i)
		if !field.CanInterface() {
			continue
		}
		if field.Kind() == reflect.Ptr && field.IsNil() {
			continue
		}
		values = append(values, field.Interface())
		args = append(args, " $"+strconv.Itoa(n))
		n += 1
		fields = append(fields, t.Field(i).Tag.Get("db"))

	}

	tableName := `"` + v.Type().Name() + `"`
	q := "INSERT INTO " + tableName + " (" + strings.Join(fields, ", ") + ") VALUES ("
	q += strings.Join(args, ", ") + ") "
	ctx := context.Background()

	fmt.Println(q)
	tx, _ := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault})
	_, execErr := tx.Exec(q, values...)
	if execErr != nil {
		if strings.Contains(execErr.Error(), "duplicate key value violates unique constraint") {
			return apiError.UserAlreadyExist
		} else {
			fmt.Println("failed to execute query: %w", execErr)
			return apiError.InternalError
		}
	}
	if err := tx.Commit(); err != nil {
		fmt.Println("failed to commit transaction: %w", execErr)
		return apiError.InternalError
	}
	return nil
}
