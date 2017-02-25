package security

import (
	"github.com/labstack/echo"
	"net/http"
	"github.com/dgrijalva/jwt-go"
	"github.com/guidola/go-utils/database"
	"time"
	"fmt"
	"gitlab.com/terno/TernoAPI/model"
	"path/filepath"
	"gopkg.in/mgo.v2/bson"
	"os"
	"github.com/labstack/gommon/log"
)


var openApiUrls = map[string]struct{}{"/login": {}, "/register": {}} //for now assuming everything in there is login related and therefore does not require jwt

//returns true if the given route is marked as public and therefore accessible without authentication
func NonAuthenticationRequired(route string) bool {
	_, ok := openApiUrls[route]
	return ok
}


func LoadAuthenticationRoutes(e *echo.Echo){

	e.POST("/login", HandleLoginRequest)
	e.GET("/logout", HandleLogoutRequest)

}

// request handler that returns a json formatted string with the response to the authentication attempt
func HandleLoginRequest(c echo.Context) error{

	u := new(model.LoginRequest)
	if err := c.Bind(u); err != nil {
		return err
	}

	if allowed, mongo_id, err := MONGO_authenticateUser(u); err == nil && allowed {
		tokenstring := JwtGetRSAToken(mongo_id)
		return c.JSON(http.StatusOK, tokenstring)
	} else {
		return c.JSON(http.StatusForbidden, nil)
	}

}


func MONGO_authenticateUser(credentials *model.LoginRequest) (bool, string, error){

	mg := database.GetMongoInstance().GetCopy()

	collection := mg.DB(model.TernoDB).C(model.UsersCollection)
	query := collection.Find(bson.M{
		"$or": []interface{}{
			bson.M{"email" : credentials.Uuid, "pwd": credentials.Pwd},
			bson.M{"username": credentials.Uuid, "pwd": credentials.Pwd},
	}})

	result, err := query.Count()

	if err != nil {
		return false, "", err
	}

	var mongo_id_struct struct{ Id  string  `json:"_id" bson:"_id"`}
	err = query.One(&mongo_id_struct)

	return result == 1, mongo_id_struct.Id, nil
}


// request handler that logouts user from the system, invalidates the associated JWT therefore logging him out
// logout request is parameterless since we are going to use the JWT to perform that action
func HandleLogoutRequest(c echo.Context) error{

	token_string, err := JwtFromHeader(echo.HeaderAuthorization)(c)
	config := DefaultJWTConfig

	if err != nil {
		return c.JSON(http.StatusBadRequest, nil)
	}

	token, err := jwt.Parse(token_string, func(t *jwt.Token) (interface{}, error) {
		// Check the signing method
		if t.Method.Alg() != config.SigningMethod {
			return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
		}
		return config.SigningKey, nil

	})

	if err != nil {
		return c.JSON(http.StatusUnauthorized, nil)
	}

	InvalidateJWT(*token)
	return c.JSON(http.StatusOK, nil)
}


// **********************
// JWT related functions
// **********************

var rsaPrivateKeyLocation, _ = filepath.Abs("private_key.pem")

//claim values
const(
	ExpirationTime = 7200 //2h //set to 30 seconds for debuging purposes, raise that to a realistic value once all is stable
	TokenIssuer = "api.terno.io"
)

//returns a signed valid JWT ready to send back to the end user
func JwtGetRSAToken(mongo_id string) (string) {

	/* create the token and configure it to work with the following header
	 *  {
	 *      "typ":"JWT"
	 *      "alg":"RS512"
	 *  }
	 */
	token := jwt.New(jwt.GetSigningMethod("RS512"))

	//define the token claims
	token.Claims.(jwt.MapClaims)["sub"] = mongo_id
	token.Claims.(jwt.MapClaims)["exp"] = time.Now().Unix() + ExpirationTime
	token.Claims.(jwt.MapClaims)["iss"] = TokenIssuer

	tokenstring, _ := token.SignedString(GetRSAPrivateKey(os.Getenv(JwtCertsLocation) + "private_key.pem"))

	return tokenstring
}


// ****************************
// Redis interaction functions
// ****************************

func InvalidateJWT(token jwt.Token){

	// add token to redis with an expiration time of the expiration time of the token to invalidate this token and not
	// allow access to the system when this token is provided

	redis := database.GetRedisInstance()
	_, err := redis.Execute("SET", token.Raw, 0, "EX", int64(token.Claims.(jwt.MapClaims)["exp"].(float64)) - time.Now().Unix())

	if err != nil {
		log.Warnf("Failed to store instance on redis")
	}

}


func IsJWTValid(token jwt.Token) bool {

	redis := database.GetRedisInstance()
	response, err := redis.Execute("EXISTS", token.Raw)

	if err != nil {
		log.Warnf("Failed check existance of token on redis")
	}

	if val, _ := response.Int(); val == 1 {
		return false
	} else{
		return true
	}

}
