package firebasehelpers

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/knq/jwt"
	"github.com/knq/jwt/gserviceaccount"
	"github.com/stretchr/testify/assert"
)

type myClaims struct {
	jwt.Claims
	Uid         string            `json:"uid"`
	ExtraClaims map[string]string `json:"claims"`
}

var serviceAccount []byte = []byte(`{
  "type": "service_account",
  "project_id": "foobar",
  "private_key_id": "f6eefe121217df59c7d901dbf549ad1fa3424",
  "private_key": "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAzAkXNREJAowREhX73nFbBFkdmsrLVHknHdjwd2X0wU8QIZcK\nvCexsiG7ZdjAS5OXrvp6QgjjQxZ8DrmCr6iXTLEaxa2ukjhuxDHkwunDYl2JqgVB\nWXN8exF5w8GmK5PglcedCe+ZaQGv0tln0XIknK9/R4Rtu7cNT5kCS0eeNcfF3zXA\nRomAwbG9/VpNImJMKIqoh5pAD/gfbqxoEwXpxN+d0CEVf73boe3JoRBQ+9J9Xk9Q\nJhtKE/jWDuBv/2K/wHLMJyfq9zRXeIAkjQRWxUdihfqr0ETtGnDlhF7BrX5VC8fF\ntIinBWSPxYVzHdj9OI2/gkyHt20VpghwWOw2wQIDAQABAoIBAFGTCN5Ek0+bZG/Q\nrkR/GZ6han6quaRqU8NRKsLx1ms7Cv4C/12+mQLZDa1ofWk59xkUN7ETEJmP8cWJ\nUcCdLPCSlluWVwdK3K5ALG/pOh6nuxRoyXnT/F7P29jyIVem5dG8XwLL8o/TBtLL\n7QAGHLEwUTjsr1qvkvjR+eLTHWPuZiw4+NFJ8xMqq/DNc2neGVIfJ27UPO5R7LIZ\nIzokDTagRi+f2klXqWHYolTqw6ycmqdKKYbDRKq/ji96gEUyf1VSHfZHCD+T6odU\nhO8YBqo+gQFzjUggnU4s5xtcQsUsHDtEOxlGHyfZkxm1ZQrajkcJ9HvGAObxPIkQ\nVSLsxeECgYEA7/4DvSDa3V2oizfr09rVihM5FFPUyJwsNKZrtsfAqGx9jqgxCSBR\nOH5RHSRkxuuf07LoK+jYEHSUyoODtLqaa1gyb9wm8p3lGCCuTad6th6HC8MEVizC\nRWJs43ig6EIpP6vfHYNkQQlKZBI0wJ3Y5M3RKAwGX2l484GuwMpcYrUCgYEA2aUY\n34C2vnaGEh/K7++fVFx7AlUmVMgz/9i/G0no5MZ6rBsyaKGbyAlPcrDLHGh6qDl+\nTwEg9OYHCvK0h4dFdjvNWijCVnMt+AwSDALvyaAJmoJHG5ZIUBJtb5tZ0frZQt75\n412LuWxG7b6a05ljLJYxRckP9w2f/0JgJtbcz10CgYAWycjzGX6OzIjnh0zWVg42\nyTJ/UqJ+1g2AhljuBzOtCng1ppTZZ/8uXRg4qy8CkHchs/hFyxtRHLDQNgK4k4t8\nK+jGJGJyYTnSu6+xYfjN+EIchM0RnbhovDrYsqicxUODbz+FXueTIV21+OCXdaWV\nvFFi+xlT0AETJjpAxjZVjQKBgBIIYdUy3vFM9LLPu4rBudvNhcudrn1b0SMjnEHw\nj8FUyJk176lHqpaaXuDL0ShbZ75EdTiqiUaBQJghn9+Sz6iKL+uGcQOkq2xf46bn\nH2L/RYxtuuKIQxmPTU3v+zMwq4uk2eOCvq7wT7gnEMDzdoodL5vumsoHcPg/UaQm\nLUlpAoGBALVAE3bcZWXjusnJA4b+JCFY+bwLTGMJczjWbPO8ql5Cw/1JdHgv4Szd\nMAYZKhAiIsK4tBMGBImustlkcHZt4FlpmkIvMPXL0RBeBbmZ0xXkbq9XlU7Qkjr4\nCasjlj1xu/20DkaycWvIRcBcjZ2QokfmW8U0RL2Lt/aDp0mYVDkL\n-----END RSA PRIVATE KEY-----",
  "client_email": "firebase-adminsdk-nlcmp@foobar.iam.gserviceaccount.com",
  "client_id": "106362299020094633032",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://accounts.google.com/o/oauth2/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/firebase-adminsdk-nlcmp%40foobar.iam.gserviceaccount.com"
}`)

var publicKey []byte = []byte("-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzAkXNREJAowREhX73nFb\nBFkdmsrLVHknHdjwd2X0wU8QIZcKvCexsiG7ZdjAS5OXrvp6QgjjQxZ8DrmCr6iX\nTLEaxa2ukjhuxDHkwunDYl2JqgVBWXN8exF5w8GmK5PglcedCe+ZaQGv0tln0XIk\nnK9/R4Rtu7cNT5kCS0eeNcfF3zXARomAwbG9/VpNImJMKIqoh5pAD/gfbqxoEwXp\nxN+d0CEVf73boe3JoRBQ+9J9Xk9QJhtKE/jWDuBv/2K/wHLMJyfq9zRXeIAkjQRW\nxUdihfqr0ETtGnDlhF7BrX5VC8fFtIinBWSPxYVzHdj9OI2/gkyHt20VpghwWOw2\nwQIDAQAB\n-----END PUBLIC KEY-----")

func TestCustomToken(t *testing.T) {
	// Test with random private key
	gsa, err := gserviceaccount.FromJSON(serviceAccount)
	if err != nil {
		panic(err)
	}
	token, err := CustomTokenFromServiceAccount(gsa, "foobar", nil)
	if err != nil {
		panic(err)
	}
	rs256, err := jwt.RS256.New(jwt.PEM{publicKey})
	cl2 := myClaims{}
	err = rs256.Decode(token, &cl2)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, "https://identitytoolkit.googleapis.com/google.identity.identitytoolkit.v1.IdentityToolkit", cl2.Audience)
	assert.Equal(t, "firebase-adminsdk-nlcmp@foobar.iam.gserviceaccount.com", cl2.Issuer)
	assert.Equal(t, "firebase-adminsdk-nlcmp@foobar.iam.gserviceaccount.com", cl2.Subject)
	assert.Equal(t, json.Number(strconv.FormatInt(time.Now().Unix(), 10)), cl2.IssuedAt)
	assert.Equal(t, json.Number(strconv.FormatInt(time.Now().Unix()+3600, 10)), cl2.Expiration)
	assert.Equal(t, "foobar", cl2.Uid)
}

func TestCustomTokenWithClaims(t *testing.T) {
	// Test with random private key
	gsa, err := gserviceaccount.FromJSON(serviceAccount)
	if err != nil {
		panic(err)
	}
	token, err := CustomTokenFromServiceAccount(gsa, "foobar", map[string]string{"foo": "bar", "fiz": "fuz"})
	if err != nil {
		panic(err)
	}
	rs256, err := jwt.RS256.New(jwt.PEM{publicKey})
	cl2 := myClaims{}
	err = rs256.Decode(token, &cl2)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, map[string]string{"fiz": "fuz", "foo": "bar"}, cl2.ExtraClaims)
}
