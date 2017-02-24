package security

import (
	"crypto/rsa"
	"io/ioutil"
	"github.com/dgrijalva/jwt-go"
	"gitlab.com/terno/TernoAPI/error_msg"
	"fmt"
)

const JwtCertsLocation = "JWT_CERTS_LOCATION"

func GetRSAPublicKey(filepath string) *rsa.PublicKey {

	key_file, err_1 := ioutil.ReadFile(filepath)
	if err_1 != nil {
		panic(error_msg.FileNotFoundError)
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
		panic(error_msg.FileNotFoundError)
	}

	priv_rsa_key, err := jwt.ParseRSAPrivateKeyFromPEM(key_file)
	if err != nil {
		panic(err)
	}
	return priv_rsa_key
}