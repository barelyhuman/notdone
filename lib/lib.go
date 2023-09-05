package lib

import (
	"crypto/rand"
	"encoding/base32"
	"text/template"
)

func Bail(err error) {
	if err != nil {
		panic(err)
	}
}

func MustParse(name string, tmpl string) *template.Template {
	t, err := template.New(name).Parse(tmpl)
	Bail(err)
	return t
}

func CryptoRandomToken(length int) string {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return base32.StdEncoding.EncodeToString(randomBytes)[:length]
}
