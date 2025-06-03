package authentication

import (
	"encoding/json"
	"fmt"
	"forum/utils"
	"io"
	"log"
	"net/http"
	"net/url"
)

// WHAT WE DO in callbackHandler functions:
// 		Step 1. Extract the code from the query string
// 		Step 2. Exchange it for an access token (via a POST request to the API)
// 		Step 3. Use the access token to get user info (name, email)
// 		Step 4. Check if the user exists in our DB and create a new session

func FacebookAuthHandler(w http.ResponseWriter, r *http.Request) {
	params := url.Values{}
	params.Add("client_id", facebookClientID)
	params.Add("redirect_uri", facebookRedirectURI)
	params.Add("response_type", "code")
	params.Add("scope", "email public_profile")

	authURL := fmt.Sprintf("%s?%s", facebookAuthEndpoint, params.Encode())

	http.Redirect(w, r, authURL, http.StatusFound)
}

func FacebookCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1:
	code := r.URL.Query().Get("code")
	if code == "" {
		utils.RenderTemplate(w, "failedoAuth.html", "Facebook")
		return
	}

	// Step 2:
	facebookToken, err := facebookTokenExchange(code)
	if err != nil {
		log.Println("Facebook token exchange failed:", err)
		utils.RenderTemplate(w, "failedoAuth.html", "Facebook")
		return
	}

	// Step 3:
	facebookUser, err := userInfoFacebookRequest(facebookToken)
	if err != nil {
		log.Println("Facebook user info request failed:", err)
		utils.RenderTemplate(w, "failedoAuth.html", "Facebook")
		return
	}

	// Step 4:
	userID, err := findOrCreateUser(facebookUser.Email, facebookUser.Name, "facebook")
	if err != nil {
		log.Println("DB error:", err)
		utils.RenderTemplate(w, "failedoAuth.html", "Facebook")
		return
	}

	createSession(userID, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

// =============================== Const variables and helper functions =====================================

const facebookClientID = "862871986031202"
const facebookClientSecret = "8e32154a9066b79957cd21beb44545d0"

const facebookAuthEndpoint = "https://www.facebook.com/v18.0/dialog/oauth"
const facebookRedirectURI = "http://localhost:8080/auth/facebook/callback"

const facebookTokenEndpoint = "https://graph.facebook.com/v18.0/oauth/access_token"
const facebookUserInfoEndpoint = "https://graph.facebook.com/me"

type FacebookTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}
type FacebookUserStruct struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// this is Step 2 helper function:
func facebookTokenExchange(code string) (FacebookTokenResponse, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", facebookClientID)
	data.Set("client_secret", facebookClientSecret)
	data.Set("redirect_uri", facebookRedirectURI)

	tokenResp, tokenRespErr := http.PostForm(facebookTokenEndpoint, data)
	if tokenRespErr != nil {
		return FacebookTokenResponse{}, fmt.Errorf("failed to post to token endpoint: %w", tokenRespErr)
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		return FacebookTokenResponse{}, fmt.Errorf("received non-OK status code: %d from user info endpoint", tokenResp.StatusCode)
	}

	body, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		return FacebookTokenResponse{}, fmt.Errorf("failed to read token response body: %w", err)
	}

	var facebookToken FacebookTokenResponse

	tokenUnmarshalErr := json.Unmarshal(body, &facebookToken)
	if tokenUnmarshalErr != nil {
		return FacebookTokenResponse{}, fmt.Errorf("failed to unmarshal token: %w", tokenUnmarshalErr)
	}

	if facebookToken.AccessToken == "" {
		return FacebookTokenResponse{}, fmt.Errorf("access token missing in response")
	}
	return facebookToken, nil
}

// this is Step 3 helper function:
func userInfoFacebookRequest(token FacebookTokenResponse) (FacebookUserStruct, error) {
	userInfoReq, userinfoReqErr := http.NewRequest("GET", facebookUserInfoEndpoint+"?fields=id,name,email", nil)
	if userinfoReqErr != nil {
		return FacebookUserStruct{}, fmt.Errorf("failed to create user info request: %w", userinfoReqErr)
	}
	userInfoReq.Header.Set("Authorization", "Bearer "+token.AccessToken)

	client := &http.Client{}
	userInfo, userInfoErr := client.Do(userInfoReq)
	if userInfoErr != nil {
		return FacebookUserStruct{}, fmt.Errorf("failed to send user info request: %w", userInfoErr)
	}
	defer userInfo.Body.Close()

	if userInfo.StatusCode != http.StatusOK {
		return FacebookUserStruct{}, fmt.Errorf("received non-OK status code: %d from user info endpoint", userInfo.StatusCode)
	}

	body, err := io.ReadAll(userInfo.Body)
	if err != nil {
		return FacebookUserStruct{}, fmt.Errorf("failed to read user info response body: %w", err)
	}

	var facebookUser FacebookUserStruct

	userInfoUnmarshalErr := json.Unmarshal(body, &facebookUser)
	if userInfoUnmarshalErr != nil {
		return FacebookUserStruct{}, fmt.Errorf("failed to unmarshal user info response or missing email: %w", userInfoUnmarshalErr)
	}

	if facebookUser.Email == "" {
		return FacebookUserStruct{}, fmt.Errorf("email missing in user info response")
	}

	return facebookUser, nil
}
