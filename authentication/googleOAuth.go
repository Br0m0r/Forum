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

// adds the query parameters to the standard Google redirect url we have above
// and redirects the user to this url
func GoogleAuthHandler(w http.ResponseWriter, r *http.Request) {
	params := url.Values{}
	params.Add("client_id", googleClientID)
	params.Add("redirect_uri", googleRedirectURI)
	params.Add("response_type", "code")
	params.Add("scope", "email profile")
	// we tell Google what we need from this user's info
	params.Add("access_type", "offline")
	//offline means that later that we get a refresh token, it doesnt expire after 1 hr
	params.Add("prompt", "consent")
	// consent forces Google to show the consent screen every time, even if the user
	// already authorized the app before. this makes sure we always get a refresh token

	authURL := fmt.Sprintf("%s?%s", googleAuthEndpoint, params.Encode())
	// Sprintf("%s?%s", ...): string(baseURL) + "?" + string(params.Encode() result)
	// Encode() takes all the url.Values and turns them into a properly encoded query string
	http.Redirect(w, r, authURL, http.StatusFound)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1:
	code := r.URL.Query().Get("code")
	if code == "" {
		utils.RenderTemplate(w, "failedoAuth.html", "Google")
		return
	}

	// Step 2:
	googleToken, err := googleTokenExchange(code)
	if err != nil {
		log.Println("Google token exchange failed:", err)
		utils.RenderTemplate(w, "failedoAuth.html", "Google")
		return
	}

	// Step 3:
	googleUser, err := userInfoGoogleRequest(googleToken)
	if err != nil {
		log.Println("Google user info request failed:", err)
		utils.RenderTemplate(w, "failedoAuth.html", "Google")
		return
	}

	// Step 4:
	userID, err := findOrCreateUser(googleUser.Email, googleUser.Name, "google")
	if err != nil {
		log.Println("DB error:", err)
		utils.RenderTemplate(w, "failedoAuth.html", "Google")
		return
	}

	createSession(userID, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

// =============================== Const variables and helper functions =====================================

const googleClientID = "12933737889-jbk18o6c8e9pa5pmb4a8psi72j1sjsm7.apps.googleusercontent.com"
const googleClientSecret = "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"

// user is redirected here when he clicks "Login with Google" along with some query parameters
const googleAuthEndpoint = "https://accounts.google.com/o/oauth2/v2/auth"

// after a successful login Google redirects the user back to this URI with
// a short-lived authorization code in the query string (?code=abc123)
const googleRedirectURI = "http://localhost:8080/auth/google/callback"

// this is the official Google OAuth 2.0 token exchange endpoint
const googleTokenEndpoint = "https://oauth2.googleapis.com/token"

// this is the official Google OAuth 2.0 user info endpoint
const googleUserInfoEndpoint = "https://www.googleapis.com/oauth2/v2/userinfo"

// struct to map the JSON response returned by Google when we exchange the code for an access token
type GoogleTokenResponse struct {
	AccessToken string `json:"access_token"`
}

// struct to hold user info when we request it from Google
type GoogleUserStruct struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// this is Step 2 helper function:
func googleTokenExchange(code string) (GoogleTokenResponse, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", googleClientID)
	data.Set("client_secret", googleClientSecret)
	data.Set("redirect_uri", googleRedirectURI)
	data.Set("grant_type", "authorization_code")
	// this key-value pair specifies what kind of token exchange we are doing

	// we send a POST request with the data above to Google token endpoint
	// if no error arises we recieve an access token
	tokenResp, tokenRespErr := http.PostForm(googleTokenEndpoint, data)
	if tokenRespErr != nil {
		return GoogleTokenResponse{}, fmt.Errorf("failed to post to token endpoint: %w", tokenRespErr)
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		// tokenResp.StatusCode is a field in the *http.Response struct (which is tokenResp)
		// http.StatusOK is just a constant with value 200 (success). If Google responds with something else (400, 401, 500),
		// it means the request failed — maybe bad client ID, code expired, etc.
		return GoogleTokenResponse{}, fmt.Errorf("received non-OK status code: %d from user info endpoint", tokenResp.StatusCode)
	}

	// the body of tokenResp is a JSON string, first we read it...
	body, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		return GoogleTokenResponse{}, fmt.Errorf("failed to read token response body: %w", err)
	}

	var googleToken GoogleTokenResponse

	// ...and then we unmarshal and save it in a variable of type: GoogleTokenResponse struct
	tokenUnmarshalErr := json.Unmarshal(body, &googleToken)
	if tokenUnmarshalErr != nil {
		return GoogleTokenResponse{}, fmt.Errorf("failed to unmarshal token: %w", tokenUnmarshalErr)
	}

	if googleToken.AccessToken == "" {
		return GoogleTokenResponse{}, fmt.Errorf("access token missing in response")
	}
	return googleToken, nil
}

// this is Step 3 helper function:
func userInfoGoogleRequest(token GoogleTokenResponse) (GoogleUserStruct, error) {
	// we create the request
	userInfoReq, userinfoReqErr := http.NewRequest("GET", googleUserInfoEndpoint, nil)
	if userinfoReqErr != nil {
		return GoogleUserStruct{}, fmt.Errorf("failed to create user info request: %w", userinfoReqErr)
	}

	// this HTTP header for our request tells the API that we are authorized, so we give it our TOKEN
	// "Bearer " is the token type. It means the token grants access to protected resources
	// (this is how OAuth-protected APIs authenticate such requests)
	userInfoReq.Header.Set("Authorization", "Bearer "+token.AccessToken)

	// we send the "GET" request we created above using a custom client and method Do()
	client := &http.Client{}
	// this is an HTTP client that makes the request, it's like a browser that can send requests and read responses
	userInfo, userInfoErr := client.Do(userInfoReq)
	if userInfoErr != nil {
		return GoogleUserStruct{}, fmt.Errorf("failed to send user info request: %w", userInfoErr)
	}
	defer userInfo.Body.Close()

	if userInfo.StatusCode != http.StatusOK {
		return GoogleUserStruct{}, fmt.Errorf("received non-OK status code: %d from user info endpoint", userInfo.StatusCode)
	}

	body, err := io.ReadAll(userInfo.Body)
	if err != nil {
		return GoogleUserStruct{}, fmt.Errorf("failed to read user info response body: %w", err)
	}

	var googleUser GoogleUserStruct

	// unmarshal and saving in a variable of type: GoogleUserStruct struct
	userInfoUnmarshalErr := json.Unmarshal(body, &googleUser)
	if userInfoUnmarshalErr != nil {
		return GoogleUserStruct{}, fmt.Errorf("failed to unmarshal user info response or missing email: %w", userInfoUnmarshalErr)
	}

	if googleUser.Email == "" {
		return GoogleUserStruct{}, fmt.Errorf("email missing in user info response")
	}

	return googleUser, nil
}
