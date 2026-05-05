package auth

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
)

const passwordSalt = "forum-minimal-salt"

func HashPassword(password string) string {
    sum := sha256.Sum256([]byte(passwordSalt + password))
    return hex.EncodeToString(sum[:])
}

func RandomToken() (string, error) {
    b := make([]byte, 32)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    return hex.EncodeToString(b), nil
}
