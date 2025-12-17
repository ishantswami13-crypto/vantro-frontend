package router

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

var jwtKey = []byte("supersecretapikey") // move to ENV later

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResp struct {
	Token string `json:"token"`
}

func (r *Router) LoginHandler(w http.ResponseWriter, req *http.Request) {
	var body LoginReq
	json.NewDecoder(req.Body).Decode(&body)

	// TEMP: Hardcode user (later add DB)
	if body.Email != "admin@vantro.com" || body.Password != "123456" {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": body.Email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})

	signed, _ := token.SignedString(jwtKey)

	json.NewEncoder(w).Encode(LoginResp{Token: signed})
}
