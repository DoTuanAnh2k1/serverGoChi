package o_auth_token

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"serverGoChi/models/config_models"
	"serverGoChi/src/log"
	"serverGoChi/src/utils/request"
)

type OauthToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    uint64 `json:"expires_in"`
	Scope        string `json:"scope"`
}

type tokenRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	GrantType string `json:"grant_type"`
}

func RequestAccessToken(user request.User) (OauthToken, error) {
	var token OauthToken
	tokenReq := tokenRequest{
		Username:  user.UserName,
		Password:  user.Password,
		GrantType: "password",
	}

	jsonData, err := json.Marshal(tokenReq)
	if err != nil {
		log.Logger.Error("Cant marshal Token Req to send req access token: ", err)
		return token, err
	}

	bodyReader := bytes.NewReader(jsonData)
	req, err := http.NewRequest(http.MethodPost, config_models.TokenUrl, bodyReader)
	if err != nil {
		log.Logger.Error("Cant create request: ", err)
		return token, err
	}

	req.Header.Add("Authorization", createAuthHeaderString("client", "secret"))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Logger.Error("Cant send req: ", err)
		return token, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Logger.Errorf("Error reading response body: %v", err)
		return token, err
	}

	err = json.Unmarshal(body, &token)
	if err != nil {
		log.Logger.Error("Cant parse body: ", err)
		return token, err
	}
	return token, nil
}

func createAuthHeaderString(username, password string) string {
	auth := username + ":" + password
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	authHeader := "Basic " + encodedAuth
	return authHeader
}
