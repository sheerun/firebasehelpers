package firebasehelpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/knq/jwt/gserviceaccount"
)

func CustomTokenFromServiceAccount(gsa *gserviceaccount.GServiceAccount, uid string, claims map[string]string) ([]byte, error) {
	var err error

	// simple check
	if gsa.ProjectID == "" || gsa.ClientEmail == "" || gsa.PrivateKey == "" {
		return nil, errors.New("google service account credentials missing project_id, client_email or private_key")
	}

	if uid == "" {
		return nil, errors.New("custom token must have non-empty uid assigned")
	}

	if len(uid) > 128 {
		return nil, errors.New("custom token uid length must be less than or equal to 128 characters")
	}

	signer, err := gsa.Signer()

	if err != nil {
		return nil, err
	}

	jwt := make(map[string]interface{})

	now := time.Now()
	n := json.Number(strconv.FormatInt(now.Unix(), 10))

	jwt["iat"] = n
	jwt["exp"] = json.Number(strconv.FormatInt(now.Add(1*time.Hour).Unix(), 10))
	jwt["aud"] = "https://identitytoolkit.googleapis.com/google.identity.identitytoolkit.v1.IdentityToolkit"
	jwt["iss"] = gsa.ClientEmail
	jwt["sub"] = gsa.ClientEmail
	jwt["uid"] = uid

	if claims != nil && len(claims) > 0 {
		jwt["claims"] = claims
	}

	// encode token
	token, err := signer.Encode(jwt)
	if err != nil {
		return nil, fmt.Errorf("jwt/bearer: could not encode claims: %v", err)
	}

	return token, nil
}
