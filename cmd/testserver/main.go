package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Login []struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"login"`
}

var serverConfig Config

func main() {
	// Load config
	data, err := os.ReadFile("configs/test_config.yaml")
	if err != nil {
		log.Printf("Warning: Could not read config file: %v. Using defaults.", err)
	} else {
		err = yaml.Unmarshal(data, &serverConfig)
		if err != nil {
			log.Printf("Warning: Could not parse config file: %v", err)
		}
	}

	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("cmd/testserver/static"))
	mux.Handle("/", fs)

	// Thecus Typical Endpoint
	mux.HandleFunc("/adm/login.php", loginHandler)

	host := flag.String("host", "", "Host to bind to (leave empty for all interfaces)")
	flag.Parse()

	port := ":8082"
	bindAddr := *host + port

	server := &http.Server{
		Addr:         bindAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	displayHost := *host
	if displayHost == "" {
		displayHost = "0.0.0.0"
	}
	fmt.Printf("--- 🌐 THECUS SIMULATOR (LAN MODE) ---\n")
	fmt.Printf("Endpoint: http://%s%s/adm/login.php\n", displayHost, port)
	fmt.Printf("Expected Fields: p_user/username, p_pass/pwd\n")
	log.Fatal(server.ListenAndServe())
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Logging for debug
	fmt.Printf("[%s] Login Attempt: ", time.Now().Format("15:04:05"))
	for k, v := range r.PostForm {
		fmt.Printf("%s=%v ", k, v)
	}
	fmt.Println()

	// Thecus field names can be multiple, we check the main ones
	username := r.FormValue("p_user")
	if username == "" {
		username = r.FormValue("username")
	}
	password := r.FormValue("p_pass")
	if password == "" {
		password = r.FormValue("pwd")
	}

	// Simulate old NAS CPU delay
	time.Sleep(100 * time.Millisecond)

	isValid := false
	for _, cred := range serverConfig.Login {
		if username == cred.Username && password == cred.Password {
			isValid = true
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if isValid {
		resp := map[string]interface{}{
			"success": true,
		}
		json.NewEncoder(w).Encode(resp)
	} else {
		resp := map[string]interface{}{
			"success": false,
			"errormsg": map[string]string{
				"title": "Authentication Failed",
				"msg":   "Invalid username or password.",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}
}
