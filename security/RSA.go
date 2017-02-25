package security

import (
	"crypto/rsa"
	"io/ioutil"
	"github.com/dgrijalva/jwt-go"
	"fmt"
	"errors"
)

const JwtCertsLocation = "JWT_CERTS_LOCATION"

func GetRSAPublicKey(filepath string) *rsa.PublicKey {

	key_file, err_1 := ioutil.ReadFile(filepath)
	if err_1 != nil {
		panic(errors.New("The provided file path does not point to an existing file"))
	}

	pub_rsa_key, err := jwt.ParseRSAPublicKeyFromPEM(key_file)
	if err != nil {
		panic(err)
	}

	return pub_rsa_key
}


func GetRSAPrivateKey(filepath string) *rsa.PrivateKey {

	key_file, err_1 := ioutil.ReadFile(filepath)
	if err_1 != nil {
		fmt.Println(filepath)
		panic(errors.New("The provided file path does not point to an existing file"))
	}

	priv_rsa_key, err := jwt.ParseRSAPrivateKeyFromPEM(key_file)
	if err != nil {
		panic(err)
	}
	return priv_rsa_key
}