package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	// Login as superuser
	loginData := map[string]string{
		"identity": "test@test.com",
		"password": "abc123_PasswordDev",
	}
	b, _ := json.Marshal(loginData)
	resp, err := http.Post("http://127.0.0.1:8092/api/collections/_superusers/auth-with-password", "application/json", bytes.NewBuffer(b))
	if err != nil {
		fmt.Println("Login error:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	
	var loginResp struct {
		Token string `json:"token"`
	}
	json.Unmarshal(body, &loginResp)
	
	if loginResp.Token == "" {
		fmt.Println("No token returned:", string(body))
		return
	}

	// Try to create user
	userData := map[string]string{
		"email": "testuser@example.com",
		"password": "password123",
		"passwordConfirm": "password123",
	}
	ub, _ := json.Marshal(userData)
	req, _ := http.NewRequest("POST", "http://127.0.0.1:8092/api/collections/users/records", bytes.NewBuffer(ub))
	req.Header.Set("Authorization", loginResp.Token)
	req.Header.Set("Content-Type", "application/json")
	
	cResp, _ := http.DefaultClient.Do(req)
	defer cResp.Body.Close()
	cBody, _ := io.ReadAll(cResp.Body)
	
	fmt.Println("Create user status:", cResp.StatusCode)
	fmt.Println("Create user response:", string(cBody))
}
