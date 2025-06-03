package utils

import (
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// Generar hash bcrypt de la contraseña
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Comparar una contraseña con su hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Clave secreta para firmar los tokens (esta en env)
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET no definido")
	}
	return []byte(secret)
}

// Clave secreta para firmar los tokens (esta en env)
// var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// Genera un token JWT para un usuario
func GenerateJWT(userID uint, rol string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,               
		"rol": rol,                 
		"iat": time.Now().Unix(),   
		"exp": time.Now().Add(time.Hour * 24).Unix(), // Expira en 24 hora (modificar, al terminar pruebas)
	})

	return token.SignedString(getJWTSecret())
	// return token.SignedString(jwtSecret)
}

// Validar token
// func ValidateJWT(tokenString string) (*jwt.Token, error) {
// 	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 		// Verifica el método de firma
// 		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, jwt.ErrSignatureInvalid
// 		}
// 		return jwtSecret, nil
// 	})
// }
// Validar token
func ValidateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return getJWTSecret(), nil
	})
}
