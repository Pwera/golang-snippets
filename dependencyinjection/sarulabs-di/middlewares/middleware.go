package middlewares

import (
	"fmt"
	"net/http"

	"github.com/pwera/di/helpers"
	"go.uber.org/zap"
)

func PanicRecoveryMiddleware(h http.HandlerFunc, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error(fmt.Sprint(rec))

				helpers.JSONResponse(w, 500, map[string]interface{}{
					"error": "Internal Error",
				})
			}

		}()
		h(w, r)
	}
}
