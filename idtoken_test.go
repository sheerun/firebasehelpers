package firebasehelpers

import (
	"testing"

	"github.com/knq/jwt/gserviceaccount"
	"github.com/stretchr/testify/assert"
	httpmock "gopkg.in/jarcoal/httpmock.v1"
)

func TestIdToken(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Test with random private key
	gsa, err := gserviceaccount.FromJSON(serviceAccount)
	if err != nil {
		panic(err)
	}
	token, err := CustomTokenFromServiceAccount(gsa, "foobar", nil)
	if err != nil {
		panic(err)
	}

	httpmock.RegisterResponder("POST", "https://www.googleapis.com/identitytoolkit/v3/relyingparty/verifyCustomToken?alt=json&key=randomkey",
		httpmock.NewStringResponder(200, `{
 "kind": "identitytoolkit#VerifyCustomTokenResponse",
 "idToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImV4cCI6MTIzNDU2Nzg5MH0.BgU93rYDuJqB5urkunSsgj1cvmfuEX6PiqPZcpkLc0k",
 "refreshToken": "foobar",
 "expiresIn": "3600"
}`))

	idToken, err := IdTokenFromCustomToken(token, "randomkey")

	if err != nil {
		panic(err)
	}

	assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImV4cCI6MTIzNDU2Nzg5MH0.BgU93rYDuJqB5urkunSsgj1cvmfuEX6PiqPZcpkLc0k", string(idToken.AccessToken))
	assert.Equal(t, "foobar", idToken.RefreshToken)
	assert.Equal(t, int64(1234567890), idToken.Expiry.Unix())
}
