package main

import (
	"fmt"
	"github.com/hashicorp/terraform-provider-scaffolding-framework/internal/dvc_oauth"
	"os"
)

func main() {
	token, err := dvc_oauth.GetAuthToken(os.Getenv("DEVCYCLE_CLIENT_ID"), os.Getenv("DEVCYCLE_CLIENT_SECRET"))
	if err != nil {
		return
	}
	fmt.Println(token.AccessToken)
}
