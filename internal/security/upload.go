package security

import (
	"errors"
	"net/http"
)

func ValidateImageFile(r *http.Request, maxSize int64) error {
	r.Body = http.MaxBytesReader(nil, r.Body, maxSize)
	if err := r.ParseMultipartForm(maxSize); err != nil {
		return errors.New("fichier trop volumineux")
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		return err
	}
	defer file.Close()

	buff := make([]byte, 512)
	file.Read(buff)
	mimeType := http.DetectContentType(buff)

	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/gif" {
		return errors.New("format d'image non supporté")
	}
	return nil
}
