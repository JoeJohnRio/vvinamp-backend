package security

import "golang.org/x/crypto/bcrypt"

const (
	bcryptCost = 13
)

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}
