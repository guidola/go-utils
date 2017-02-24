package security

import (
	"testing"
	"github.com/labstack/echo"
	"bytes"
	"encoding/json"
	"net/http"
	"gitlab.com/terno/TernoAPI/database"
	"os"
	"net/http/httptest"
)


func TestNonAuthenticationRequired(t *testing.T) {

	var test_cases = []struct{
		url string
		open bool
	}{
		{"/login", true},
		{"/", false},
	}

	for _, test_case := range test_cases {
		if ret := NonAuthenticationRequired(test_case.url); ret != test_case.open {
			t.Errorf("URL is %s when it should be %s", ret, test_case.open);
		}
	}

}


func TestLoginLogoutLifecycle(t *testing.T) {

	//initi redis
	var redisURI = os.Getenv(database.RedisURI)
	database.GetRedisInstance().Create("tcp", redisURI, 5)
	defer database.GetRedisInstance().Destroy()

	var mongoHosts = []string{os.Getenv(database.MongoURI)}
	database.GetMongoInstance().Create(mongoHosts)
	defer database.GetMongoInstance().Destroy()

	// make login request
	e := echo.New()
	e.Use(RSA_JWT(GetRSAPublicKey(os.Getenv(JwtCertsLocation) + "public_key.pem")))
	test_handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "foo bar")
	}
	e.GET("/testerino-divino", test_handler)
	LoadAuthenticationRoutes(e)

	payload := bytes.NewBufferString("{\"uuid\":\"patata\",\"pwd\":\"pwned\"}")
	req, _ := http.NewRequest(echo.POST, "/login", payload)
	req.Header().Set(echo.HeaderContentType, "application/json")
	res := httptest.NewRecorder()
	e.ServeHTTP(req, res)

	if res.Code != http.StatusForbidden {
		t.Errorf("Expect to get StatusForbidden when login with bad credentials but got %d", res.Code)
	}

	//login with a valid user
	payload = bytes.NewBufferString("{\"uuid\":\"test\",\"pwd\":\"pwd\"}")
	req, _ = http.NewRequest(echo.POST, "/login", payload)
	req.Header().Set(echo.HeaderContentType, "application/json")
	res = httptest.NewRecorder()
	e.ServeHTTP(req, res)

	//check we got a 200 response code with a 512bits length token
	var token_string string
	json.Unmarshal(res.Body.Bytes(), &token_string)
	if res.Code != http.StatusOK {
		t.Errorf("Expected status code to be 200 and got %d", res.Code)
	}

	//cannot check token length since it depends on the length of variable parameters on the token

	// at this point we should have a valid token and be able to perform querys but first we gotta add the jwt middleware
	// and register some route

	req, _ = http.NewRequest(echo.GET, "/testerino-divino", bytes.NewBufferString(""))
	req.Header().Set(echo.HeaderAuthorization, "jwt$" + token_string)

	res = httptest.NewRecorder()
	e.ServeHTTP(req, res)

	//we expect to get a 200 Status OK
	if res.Code != http.StatusOK {
		t.Errorf("Expected to get response code %d for /testerino-divino and got %d", http.StatusOK, res.Code)
		t.FailNow()
	}

	//if we could perform a request we can proceed to check if logout functionality works as well
	req, _ = http.NewRequest(echo.GET, "/logout", nil)
	req.Header().Set(echo.HeaderAuthorization, "jwt$" + token_string)
	res = httptest.NewRecorder()
	e.ServeHTTP(req, res)

	if res.Code != http.StatusOK {
		t.Errorf("Expected to get response code %d for /logout and got %d", http.StatusOK, res.Code)
		t.FailNow()
	}

	//at this point we should not be able to perform protected petitions anymore
	req, _ = http.NewRequest(echo.GET, "/testerino-divino", nil)
	req.Header().Set(echo.HeaderAuthorization, "jwt$" + token_string)
	res = httptest.NewRecorder()
	e.ServeHTTP(req, res)

	//we expect to get a 401 Status Unauthorized
	if res != http.StatusUnauthorized {
		t.Errorf("Expected to get response code %d for /testerino-divino and got %d", http.StatusUnauthorized, res.Code)
		t.FailNow()
	}

}