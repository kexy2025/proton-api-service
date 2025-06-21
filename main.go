package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/gorilla/mux"
)

type CheckEmailsResponse struct {
    Success     bool     `json:"success"`
    Message     string   `json:"message"`
    EmailCount  int      `json:"email_count"`
}

func main() {
    r := mux.NewRouter()
    
    r.HandleFunc("/health", healthCheck).Methods("GET")
    r.HandleFunc("/check-emails", checkEmails).Methods("POST")
    
    port := os.Getenv("PORT")
    if port == "" {
        port = "8081"
    }
    
    fmt.Printf("🚀 ProtonMail API Service starting on :%s\n", port)
    log.Fatal(http.ListenAndServe(":"+port, r))
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
        "service": "proton-api-service",
        "timestamp": time.Now().Format(time.RFC3339),
    })
}

func checkEmails(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    // Temporary mock response while we fix the ProtonMail API
    response := CheckEmailsResponse{
        Success:    true,
        Message:    "Service running - ProtonMail integration coming soon",
        EmailCount: 0,
    }
    
    json.NewEncoder(w).Encode(response)
}
