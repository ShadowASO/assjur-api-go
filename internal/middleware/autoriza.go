/*
---------------------------------------------------------------------------------------
File: autoriza.go
Autor: Aldenor
Data: 07-06-2026
Alteração:
--------------------------------------------------------------------------------------
Finalidade: Middleware para Verificar se o requisitante possui autorização para uma determinada rota.
---------------------------------------------------------------------------------------
*/
package middleware

import (
	"net/http"
	"ocrserver/internal/utils/mslogger"
	"ocrserver/internal/utils/msresponse"
	"slices"

	"github.com/gin-gonic/gin"
)

func AuthorizaMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {

		roleVal, ok := c.Get("userRole")
		if !ok {

			msresponse.Fail(
				c,
				http.StatusUnauthorized,
				"Usuário não autenticado",
				msresponse.ErrorNaoAutorizado,
				"Usuário não autenticado",
			)
			c.Abort()
			return
		}
		role, _ := roleVal.(string)

		// Admin sempre pode
		if role == "admin" || slices.Contains(allowedRoles, role) {
			c.Next()
			return
		}

		mslogger.LoggerGlobal.Infof("Acesso negado: role=%q precisa de %v", role, allowedRoles)

		msresponse.Fail(
			c,
			http.StatusForbidden,
			"Usuário sem permissão suficiente para esta ação",
			msresponse.ErrorNaoAutorizado,
			"Usuário sem permissão suficiente para esta ação",
		)
		c.Abort()
	}
}
