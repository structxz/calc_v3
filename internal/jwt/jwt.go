package jwt

import (
	// "distributed_calculator/internal/constants"
	"distributed_calculator/internal/logger"
	// "errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func MakeJWT(logger *logger.Logger, login string) (string, error) {
	secret, _ := os.LookupEnv("JWT_SECRET")
	// if !exists {
	//     secret = "6ZwDd3oHWlRj1zjayi6X9M7bAtTY7nX9sifCq3oR/9U="
	// }

	claims := jwt.MapClaims{
		"sub": login,
		"exp": time.Now().Add(24 * time.Hour).Unix(), // токен живёт 24 часа
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// func VerifyJWT(token string) bool {
// 	const hmacSampleSecret = "super_secret_signature"
// 	tokenFromString, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
// 		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 			panic(fmt.Errorf("unexpected signing method: %v", token.Header["alg"]))
// 		}

// 		return []byte(hmacSampleSecret), nil
// 	})

// 	if err != nil {
// 		logger.Error()
// 	}

// 	if claims, ok := tokenFromString.Claims.(jwt.MapClaims); ok {
// 		fmt.Println("user name: ", claims["name"])
// 	} else {
// 		panic(err)
// 	}
// }
