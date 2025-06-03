package middlewares

import (
	"net/http"

	respuestas "github.com/Ilimm9/CMedicas/Respuestas"
	"github.com/Ilimm9/CMedicas/clave"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			respuestas.RespondError(c, http.StatusUnauthorized, "Se requiere token de autenticación")
			c.Abort()
			return
		}

		// Eliminar el prefijo "Bearer " si está presente
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		token, err := clave.ValidateJWT(tokenString)
		if err != nil {
			respuestas.RespondError(c, http.StatusUnauthorized, "Token inválido: "+err.Error())
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Guardar información del usuario en el contexto
			c.Set("userID", claims["sub"])
			c.Set("userRol", claims["rol"])
			c.Next()
		} else {
			respuestas.RespondError(c, http.StatusUnauthorized, "Token inválido")
			c.Abort()
		}
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		rol := c.GetString("userRol")
		if rol != "administrador" {
			respuestas.RespondError(c, http.StatusForbidden, "Acceso restringido a administradores")
			c.Abort()
			return
		}
		c.Next()
	}
}
