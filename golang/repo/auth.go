package repo

import (
	"context"
	"database/sql"
	"time"

	"github.com/kolakdd/cache_storage/golang/apiError"
	"github.com/kolakdd/cache_storage/golang/models"
	"github.com/kolakdd/cache_storage/golang/utils"
	"github.com/redis/go-redis/v9"
	uuid "github.com/satori/go.uuid"
)

type AuthRepo interface {
	CreateUser(dto *models.RegisterDto) (*models.RegisterUserRequest, *apiError.BackendErrorInternal)
	GetUserID(dto *models.AuthDto) (*uuid.UUID, *apiError.BackendErrorInternal)
	CreateAuthToken(id *uuid.UUID) (*string, *apiError.BackendErrorInternal)
	GetByTokenAndUnauthorize(token string, del bool) (*uuid.UUID, *apiError.BackendErrorInternal)
}

type authRepo struct {
	db    *sql.DB
	cache *redis.Client
}

func NewAuthRepo(db *sql.DB, cache *redis.Client) AuthRepo {
	return &authRepo{db, cache}
}

func (r *authRepo) CreateUser(dto *models.RegisterDto) (*models.RegisterUserRequest, *apiError.BackendErrorInternal) {
	userDB := models.NewUserDB(dto.Login, dto.Pswd)
	if err := CRUDCreate(r.db, userDB); err != nil {
		return nil, err
	}
	res := &models.RegisterUserRequest{Login: userDB.Login}
	return res, nil
}

func (r *authRepo) GetUserID(dto *models.AuthDto) (*uuid.UUID, *apiError.BackendErrorInternal) {
	sqlStatement := `SELECT id, hash_password FROM "User" WHERE login=$1;`

	var id *uuid.UUID
	var hashPassword string

	row := r.db.QueryRow(sqlStatement, dto.Login)

	switch err := row.Scan(&id, &hashPassword); err {
	case sql.ErrNoRows:
		return nil, apiError.NotFound
	case nil:
		return id, nil
	default:
		return nil, apiError.InternalError
	}
}

// CreateAuthToken Создает токен авторизации в кеше
// формат токена auth: ключ 				   -   значение
// формат токена auth:{сгенерированный токен}  -  {id пользователя }
func (r *authRepo) CreateAuthToken(id *uuid.UUID) (*string, *apiError.BackendErrorInternal) {
	token := utils.GenerateToken(16)

	key := "auth:" + token
	val := id.String()
	exp := time.Duration(time.Hour * 24 * 15)

	ctx := context.Background()
	ans := r.cache.SetEx(ctx, key, val, exp)

	if ans.Err() != nil {
		return nil, apiError.RedisError
	}
	return &token, nil
}

func (r *authRepo) GetByTokenAndUnauthorize(token string, del bool) (*uuid.UUID, *apiError.BackendErrorInternal) {
	ctx := context.Background()
	var ans *redis.StringCmd
	if del {
		ans = r.cache.GetDel(ctx, "auth:"+token)
	} else {
		ans = r.cache.Get(ctx, "auth:"+token)
	}
	if ans.Err() != nil {
		return nil, apiError.Unauthorized
	}
	id, _ := uuid.FromString(ans.Val())
	return &id, nil
}
