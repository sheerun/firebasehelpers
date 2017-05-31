package firebasehelpers

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	j "github.com/bitly/go-simplejson"
	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi/transport"
	identitytoolkit "google.golang.org/api/identitytoolkit/v3"
)

func decodeSegment(seg string) ([]byte, error) {
	if l := len(seg) % 4; l > 0 {
		seg += strings.Repeat("=", 4-l)
	}

	return base64.URLEncoding.DecodeString(seg)
}

func IdTokenFromCustomToken(token []byte, apiKey string) (*oauth2.Token, error) {
	client := &http.Client{
		Transport: &transport.APIKey{Key: apiKey},
	}

	identitytoolkitService, err := identitytoolkit.New(client)

	if err != nil {
		return nil, err
	}

	request := &identitytoolkit.IdentitytoolkitRelyingpartyVerifyCustomTokenRequest{
		ReturnSecureToken: true,
		Token:             string(token),
	}

	call := identitytoolkitService.Relyingparty.VerifyCustomToken(request)

	response, err := call.Do()

	if err != nil {
		return nil, fmt.Errorf("could not verify custom token: %v", err)
	}

	ret := &oauth2.Token{
		AccessToken:  response.IdToken,
		RefreshToken: response.RefreshToken,
		TokenType:    "Bearer",
	}

	content := strings.Split(response.IdToken, ".")[1]

	data, err := decodeSegment(content)

	if err != nil {
		return nil, fmt.Errorf("cannot parse token: %v", err)
	}

	payload, err := j.NewJson(data)

	if err != nil {
		return nil, fmt.Errorf("cannot parse jwt json: %v", err)
	}

	exp, err := payload.GetPath("exp").Int64()

	if err != nil {
		return nil, fmt.Errorf("cannot fetch exp from jwt token: %v", err)
	}

	log.Printf("%d", exp)

	ret.Expiry = time.Unix(exp, 0)

	return ret, nil
}
