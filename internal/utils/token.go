package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

type LoginRequest struct {
	Username          string `json:"username"`
	Password          string `json:"password"`
	LoginProviderName string `json:"loginProviderName"`
}

type TokenDetails struct {
	Token            string `json:"token"`
	Name             string `json:"name"`
	UserName         string `json:"userName"`
	AuthProviderName string `json:"authProviderName"`
	User             User   `json:"user"`
	Timeout          int    `json:"timeout"`
	StartTime        string `json:"startTime"`
	Address          string `json:"address"`
	Partition        string `json:"partition"`
	Generation       int    `json:"generation"`
	LastUpdateMicros int64  `json:"lastUpdateMicros"`
	ExpirationMicros int64  `json:"expirationMicros"`
	Kind             string `json:"kind"`
	SelfLink         string `json:"selfLink"`
}
type LoginReference struct {
	Link string `json:"link"`
}
type User struct {
	Link string `json:"link"`
}

type LoginResponse struct {
	Token             TokenDetails   `json:"token"`
	UserName          string         `json:"username"`
	LoginReference    LoginReference `json:"loginReference"`
	LoginProviderName string         `json:"loginProviderName"`
	Generation        int            `json:"generation"`
	LastUpdateMicros  int64          `json:"lastUpdateMicros"`
}

func GetTokenFromF5(url, username, password string, insecure bool, timeout time.Duration) (TokenDetails, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	client := &http.Client{Transport: tr, Timeout: timeout * time.Second}
	loginURL := url + "/mgmt/shared/authn/login"
	loginPayload := LoginRequest{
		Username:          username,
		Password:          password,
		LoginProviderName: "tmos",
	}

	jsonData, err := json.Marshal(loginPayload)
	if err != nil {
		return TokenDetails{}, fmt.Errorf("error marshaling login payload: %w", err)
	}

	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return TokenDetails{}, fmt.Errorf("error creating login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return TokenDetails{}, fmt.Errorf("login request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return TokenDetails{}, fmt.Errorf("error reading login response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return TokenDetails{}, fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(bodyBytes, &loginResp); err != nil {
		return TokenDetails{}, fmt.Errorf("error decoding login response: %w", err)
	}

	log.Printf("New authentication token obtained from %s (expires in %d seconds)", url, loginResp.Token.Timeout)

	return loginResp.Token, nil
}
