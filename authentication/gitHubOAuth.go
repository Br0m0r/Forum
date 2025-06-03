package authentication

import (
	"encoding/json"
	"fmt"
	"forum/utils"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// WHAT WE DO in callbackHandler functions:
// 		Step 1. Extract the code from the query string
// 		Step 2. Exchange it for an access token (via a POST request to the API)
// 		Step 3. Use the access token to get user info (name, email)
// 		Step 4. Check if the user exists in our DB and create a new session

func GitHubAuthHandler(w http.ResponseWriter, r *http.Request) {
	params := url.Values{}
	params.Add("client_id", gitHubClientID)
	params.Add("redirect_uri", gitHubRedirectURI)
	params.Add("scope", "read:user user:email")

	authURL := fmt.Sprintf("%s?%s", gitHubAuthEndpoint, params.Encode())

	http.Redirect(w, r, authURL, http.StatusFound)
}

func GitHubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1:
	code := r.URL.Query().Get("code")
	if code == "" {
		utils.RenderTemplate(w, "failedoAuth.html", "GitHub1")
		return
	}

	// Step 2:
	gitHubToken, err := gitHubTokenExchange(code)
	if err != nil {
		log.Println("GitHub token exchange failed:", err)
		utils.RenderTemplate(w, "failedoAuth.html", "GitHub2")
		return
	}

	// Step 3:
	gitHubUser, err := userInfoGitHubRequest(gitHubToken)
	if err != nil {
		log.Println("GitHub user info request failed:", err)
		utils.RenderTemplate(w, "failedoAuth.html", "GitHub3")
		return
	}

	// Step 4:
	userID, err := findOrCreateUser(gitHubUser.Email, gitHubUser.Name, "github")
	if err != nil {
		log.Println("DB error:", err)
		utils.RenderTemplate(w, "failedoAuth.html", "GitHub4")
		return
	}

	createSession(userID, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

// =============================== Const variables and helper functions =====================================

const gitHubClientID = "Ov23liNRxdC40HZWyuhI"
const gitHubClientSecret = "3a4345d7c6e3a81068bb238d1775d00c4ffc4a04"

const gitHubAuthEndpoint = "https://github.com/login/oauth/authorize"
const gitHubRedirectURI = "http://localhost:8080/auth/github/callback"

const gitHubTokenEndpoint = "https://github.com/login/oauth/access_token"
const gitHubUserInfoEndpoint = "https://api.github.com/user"

type GithubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}
type GithubUserStruct struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// this is Step 2 helper function:
func gitHubTokenExchange(code string) (GithubTokenResponse, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", gitHubClientID)
	data.Set("client_secret", gitHubClientSecret)
	data.Set("redirect_uri", gitHubRedirectURI)

	tokenReq, tokenReqErr := http.NewRequest("POST", gitHubTokenEndpoint, strings.NewReader(data.Encode()))
	if tokenReqErr != nil {
		return GithubTokenResponse{}, fmt.Errorf("failed to create the request for the token endpoint: %w", tokenReqErr)
	}
	tokenReq.Header.Set("Accept", "application/json")
	// tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	tokenResp, tokenRespErr := client.Do(tokenReq)
	if tokenRespErr != nil {
		return GithubTokenResponse{}, fmt.Errorf("failed to send request to the token endpoint: %w", tokenRespErr)
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		return GithubTokenResponse{}, fmt.Errorf("received non-OK status code: %d from user info endpoint", tokenResp.StatusCode)
	}

	body, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		return GithubTokenResponse{}, fmt.Errorf("failed to read token response body: %w", err)
	}

	var gitHubToken GithubTokenResponse

	tokenUnmarshalErr := json.Unmarshal(body, &gitHubToken)
	if tokenUnmarshalErr != nil {
		return GithubTokenResponse{}, fmt.Errorf("failed to unmarshal token: %w", tokenUnmarshalErr)
	}

	if gitHubToken.AccessToken == "" {
		return GithubTokenResponse{}, fmt.Errorf("access token missing in response")
	}
	return gitHubToken, nil
}

// this is Step 3 helper function:
func userInfoGitHubRequest(token GithubTokenResponse) (GithubUserStruct, error) {
	// https://api.github.com/user/emails
	userInfoReq, userinfoReqErr := http.NewRequest("GET", gitHubUserInfoEndpoint, nil)
	if userinfoReqErr != nil {
		return GithubUserStruct{}, fmt.Errorf("failed to create user info request: %w", userinfoReqErr)
	}
	userInfoReq.Header.Set("Authorization", "Bearer "+token.AccessToken)
	userInfoReq.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	userInfo, userInfoErr := client.Do(userInfoReq)
	if userInfoErr != nil {
		return GithubUserStruct{}, fmt.Errorf("failed to send user info request: %w", userInfoErr)
	}
	defer userInfo.Body.Close()

	if userInfo.StatusCode != http.StatusOK {
		return GithubUserStruct{}, fmt.Errorf("received non-OK status code: %d from user info endpoint", userInfo.StatusCode)
	}

	body, err := io.ReadAll(userInfo.Body)
	if err != nil {
		return GithubUserStruct{}, fmt.Errorf("failed to read user info response body: %w", err)
	}

	var gitHubUser GithubUserStruct

	userInfoUnmarshalErr := json.Unmarshal(body, &gitHubUser)
	if userInfoUnmarshalErr != nil {
		return GithubUserStruct{}, fmt.Errorf("failed to unmarshal user info response or missing email: %w", userInfoUnmarshalErr)
	}

	if gitHubUser.Email == "" {
		userEmail, err := gitHubUserEmail(token)
		if err != nil {
			return GithubUserStruct{}, fmt.Errorf("failed to get user email: %w", err)
		}
		gitHubUser.Email = userEmail
	}

	return gitHubUser, nil
}

// this is a helper function for getting user's email in Step 3:
func gitHubUserEmail(token GithubTokenResponse) (string, error) {
	emailReq, emailReqErr := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if emailReqErr != nil {
		return "", fmt.Errorf("failed to create the request for the token endpoint: %w", emailReqErr)
	}
	emailReq.Header.Set("Authorization", "Bearer "+token.AccessToken)
	emailReq.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	emailResp, err := client.Do(emailReq)
	if err != nil {
		return "", fmt.Errorf("failed to request user emails: %w", err)
	}

	defer emailResp.Body.Close()

	if emailResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-OK status fetching emails: %d", emailResp.StatusCode)
	}

	// emailsList contains all the user's emails, that will or will not be markded as primary and/or verified
	var emailsList []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	err = json.NewDecoder(emailResp.Body).Decode(&emailsList)
	if err != nil {
		return "", fmt.Errorf("failed to decode email list: %w", err)
	}

	// we iterate over all the user's emails and when we find the one that is both primary and verified
	// we return this Email-string
	for _, email := range emailsList {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	return "", fmt.Errorf("no verified primary email found")
}
