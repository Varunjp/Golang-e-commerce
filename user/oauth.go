package user

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var redirectUrl string
var ClientID	string
var ClientSecret string
var GoogleOauthConfig *oauth2.Config

func GetValues(){
	redirectUrl = os.Getenv("GoogleRedirectURL")
	ClientID = os.Getenv("ClientID")
	ClientSecret = os.Getenv("ClientSecret")

	GoogleOauthConfig = &oauth2.Config{
		RedirectURL:  redirectUrl,
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
		Scopes:       []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	
}


