package security

import (
	"testing"
	"github.com/labstack/echo"
	"errors"
	"net/http"
	"bytes"
	"gitlab.com/terno/TernoAPI/database"
	"path/filepath"
	"os"
	"net/http/httptest"
)

func TestJwtFromHeader(t* testing.T){

	prefix := echo.HeaderAuthorization

	var test_cases = []struct{
		in, out string
		err error
	}{
		{"jwt$whatever", "whatever", nil},
		{"jwt$", "", errors.New("empty or invalid jwt in authorization header")},
		{"test$wathever", "", errors.New("empty or invalid jwt in authorization header")},
	}

	e := echo.New()
	req:= httptest.NewRequest(echo.GET, "/", nil)
	res:= httptest.NewRecorder()
	c := e.NewContext(req, res)

	for _, test_case := range test_cases {

		req.Header().Set(echo.HeaderAuthorization, test_case.in)

		extractor := JwtFromHeader(prefix)
		auth, err := extractor(c)
		if auth != test_case.out{
			t.Errorf("expected auth to be  '%s' and got '%s'", test_case.out, auth)
		}
		if err == nil && test_case.err != nil{
			t.Errorf("expected err to be '%s' but got nil instead", test_case.err)
		}else if test_case.err != nil{
			if err.Error() != test_case.err.Error() {
				t.Errorf("expected error to be  '%s' and got '%s'", test_case.err.Error(), err.Error())
			}
		}
	}
}

var rsaPublicKeyLocation, _ = filepath.Abs("public_key.pem")

func TestRSA_JWT(t *testing.T) {


	/* The test cases are on the following order:
	 *      · expired token
	 *      · tampered token
	 *      · non expected signing method. Checking protection against known vulnerabilities of the used library
	 */
	var must_reject_test_cases = []struct{
		token   string
		code    int
	}{
		{"eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0NjU3ODA4MjYsImlzcyI6ImFwaS50ZXJub2FwcC5jb20iLCJzdWIiOiJ0ZXN0dXNlciJ9.taaPPBZWmvvUkf2TOhrXXPaiKpb8n2M76zKw_KEmsadNL4TmE490A4VqJAaXEsjF5HjA_Q4qf4RYR34zvcndP04vNu6B86XGw755QxmufulHowrta0dKuUSztmIsG9YnyJ2VmMUBZ2l2_1udhq-4bPZB-5rkHR9KinqtlZbbCjDSYEU-FJZ80dmwQsPj2qo6FKa2v9JnyfNAWpc7ZvSRz2-nqKM8vp9GT_4-rZZUPPEO9LVhAqTd-nj4rKt7EXImhNubfYSNKhVI_YuR_C1alCe1G4TLZXZegfFMawudRVeLSoXcz6hfPoI-Oe8KQvIYWtqEpAoBPQ5wWG6L4G21dw",
		 401},
		{"eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0NjYyNzM0MjIsImlzcyI6ImFwaS50ZXJub2FwcC5jb20iLCJzdWIiOiJwYXRhdGEifQ.eiePwuyFhs3UinbFdkdsAkidFCjfhwWzgHb4tRxfZQ7YyJK9Ey4nHV4CcV04CjayptLjDO4JWj7hwJzO00znzx32e09Lt4VswdTZzlpIAf9yk8VPyChft2urzmy2NV5v5iVN2iSTCFzWgmxf90o1n-RyFU7ig5kk1edZTdTtZ2KOfxVRjMkMINKzBP2Gb0t7I8CIxvCHaT2J4ojFcNDog44pCFCX4UmB6MeCs9IjPjRv1Ge9QnfF9_XHarVRu9GnJRbL1PyRNfJq8lJ9n_ZpChdg_Fh5KifYI9cP_zxA2xtBH8amTQG1xTVCIgqZu6E3P-YQLT4oTGCD5lP4OiCdew",
		 401},
		{"eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJleHAiOjE0NjYyNzM0MjEsImlzcyI6ImFwaS50ZXJub2FwcC5jb20iLCJzdWIiOiJwYXRhdGEifQ.eiePwuyFhs3UinbFdkdsAkidFCjfhwWzgHb4tRxfZQ7YyJK9Ey4nHV4CcV04CjayptLjDO4JWj7hwJzO00znzx32e09Lt4VswdTZzlpIAf9yk8VPyChft2urzmy2NV5v5iVN2iSTCFzWgmxf90o1n-RyFU7ig5kk1edZTdTtZ2KOfxVRjMkMINKzBP2Gb0t7I8CIxvCHaT2J4ojFcNDog44pCFCX4UmB6MeCs9IjPjRv1Ge9QnfF9_XHarVRu9GnJRbL1PyRNfJq8lJ9n_ZpChPg_Fh5KifYI9cP_zxA2xtBH8amTQG1xTVCIgqZu6E3P-YQLT4oTGCD5lP4OiCdew",
		 401},
	}

	var must_pass_test_cases = []struct{
		token string
		code int

	}{
		{"eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEwMDAwMDAwMDE0NjYyNzQwOTMsImlzcyI6ImFwaS50ZXJub2FwcC5jb20iLCJzdWIiOiJwYXRhdGEifQ.fwjwCHNPJ74QBQ0PzcqMAeKAXH34HE1JiTIe4_Z89aS92Toeu3jpSHVsBvKCddyUG4yRF5dEINNFh-y46Kj6BaV8t9s8bZOiL4A-8LOixoAnYA7HrDCa1xUkN0BE7uZfLE7qybYJmgvsa6za3RVWZcUFwMd3YZa4etnKXvRcHl0MRpZspH5SvNO7c-_2o2UbJ1bXDjePRtD9xvFtaEaY0GKek_FHS9HsyKLXmKDa94pS0UeTq_jZ6_XPBGjMnp50iSbGrQPi0ukY5euimtwF6YOltWOks2xUfVD13uB3HvuPLBNYqvVZbiA0nv5F0Ke801TOpLJfDaR874dD2oGOIA",
			200},
	}

	var redisURI = os.Getenv(database.RedisURI)

	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	database.GetRedisInstance().Create("tcp", redisURI, database.RedisMaxPoolSize)
	defer database.GetRedisInstance().Destroy()

	//testing incoming tokens that must fail
	for _, test_case := range must_reject_test_cases{

		auth := "jwt$" + test_case.token
		req.Header().Set(echo.HeaderAuthorization, auth)
		m := RSA_JWT(GetRSAPublicKey(os.Getenv(JwtCertsLocation) + "public_key.pem"))(handler)
		me := m(c).(*echo.HTTPError)
		if me.Code != test_case.code {
			t.Errorf("expected HTTP error code %d and got %d instead", test_case.code, me.Code)
		}

	}

	//testing incoming tokens that must pass
	for _, test_case := range must_pass_test_cases{

		auth := "jwt$" + test_case.token
		req.Header().Set(echo.HeaderAuthorization, auth)
		m := RSA_JWT(GetRSAPublicKey(os.Getenv(JwtCertsLocation) + "public_key.pem"))(handler)
		err := m(c)
		if err != nil {
			t.Errorf("expected err to be nil and got %s instead", err)
		}

	}

	//testing that open query bypasses this middleware's handler
	e2 := echo.New()
	req2 := httptest.NewRequest(echo.POST, "/login", bytes.NewBufferString("{\"uuid\":\"patata\",\"pwd\":\"pwned\"}"))
	res2 := httptest.NewRecorder()
	c2 := e2.NewContext(req2, res2)
	handler2 := func(c2 echo.Context) error {
		return c2.String(http.StatusOK, "testopenauth")
	}
	m := RSA_JWT(GetRSAPublicKey(os.Getenv(JwtCertsLocation) + "public_key.pem"))(handler2)
	err := m(c2)
	if err != nil {
		t.Errorf("expected authentication to be bypassed and therefore err to be nil and got %s instead", err)
	}

}
