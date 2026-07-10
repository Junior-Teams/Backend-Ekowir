package utils

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

const genericServerError = "Terjadi kesalahan pada server, silahkan coba lagi nanti"

// uniqueConstraintMessages maps Postgres unique constraint names to messages
// that are safe to show on the frontend.
var uniqueConstraintMessages = map[string]string{
	"uni_users_username":  "Username sudah digunakan, silahkan gunakan username lain",
	"uni_users_email":     "Email sudah terdaftar, silahkan login",
	"idx_users_google_id": "Akun Google ini sudah terdaftar, silahkan login",
}

// RespondDBError writes a user-friendly JSON error response for a database error,
// translating raw GORM/Postgres errors (unique constraint violations, record-not-found)
// into messages safe to show on the frontend. notFoundMessage lets the caller tailor
// what's shown when the record doesn't exist. The real error is logged server-side.
func RespondDBError(context *gin.Context, err error, notFoundMessage string) {
	if err == nil {
		return
	}
	log.Println("db error:", err)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		context.JSON(http.StatusNotFound, gin.H{"error": notFoundMessage})
		context.Abort()
		return
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		message, ok := uniqueConstraintMessages[pgErr.ConstraintName]
		if !ok {
			message = "Data sudah ada, silahkan periksa kembali"
		}
		context.JSON(http.StatusConflict, gin.H{"error": message})
		context.Abort()
		return
	}

	context.JSON(http.StatusInternalServerError, gin.H{"error": genericServerError})
	context.Abort()
}

// RespondValidationError writes a friendly 400 response for an invalid request body,
// hiding raw Go binding/validation error text from the client.
func RespondValidationError(context *gin.Context, message string) {
	context.JSON(http.StatusBadRequest, gin.H{"error": message})
	context.Abort()
}

// RespondServerError logs the real error and writes a generic friendly 500 response.
func RespondServerError(context *gin.Context, err error) {
	log.Println("internal error:", err)
	context.JSON(http.StatusInternalServerError, gin.H{"error": genericServerError})
	context.Abort()
}
