package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/ProtonMail/go-proton-api"
    "github.com/gorilla/mux"
)

type EmailAccount struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type CheckEmailsRequest struct {
    Account EmailAccount `json:"account"`
}

type CheckEmailsResponse struct {
    Success     bool     `json:"success"`
    Message     string   `json:"message"`
    EmailCount  int      `json:"email_count"`
    Errors      []string `json:"errors,omitempty"`
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
    
    var req CheckEmailsRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    log.Printf("📧 Checking emails for: %s", req.Account.Email)
    
    // Initialize ProtonMail client
    manager := proton.New(
        proton.WithHostURL("https://mail.proton.me"),
        proton.WithAppVersion("web-mail@4.0.0"),
    )
    
    ctx := context.Background()
    
    // Create client
    client, auth, err := manager.NewClientWithLogin(ctx, req.Account.Email, []byte(req.Account.Password))
    if err != nil {
        log.Printf("❌ Login failed for %s: %v", req.Account.Email, err)
        json.NewEncoder(w).Encode(CheckEmailsResponse{
            Success: false,
            Message: fmt.Sprintf("Login failed: %v", err),
        })
        return
    }
    
    defer client.Close()
    
    // Get messages
    messages, err := client.GetMessages(ctx, proton.GetMessagesReq{
        Page:     0,
        PageSize: 50,
        Filter: proton.MessageFilter{
            Unread: proton.Bool(true),
        },
    })
    
    if err != nil {
        log.Printf("❌ Failed to get messages: %v", err)
        json.NewEncoder(w).Encode(CheckEmailsResponse{
            Success: false,
            Message: fmt.Sprintf("Failed to get messages: %v", err),
        })
        return
    }
    
    log.Printf("✅ Found %d unread messages", len(messages))
    
    // Mark messages as read
    processedCount := 0
    var errors []string
    
    for _, message := range messages {
        err := client.MarkMessagesRead(ctx, []string{message.ID})
        if err != nil {
            errors = append(errors, fmt.Sprintf("Failed to mark message as read: %v", err))
            continue
        }
        processedCount++
        
        log.Printf("📨 Processed: %s", message.Subject)
    }
    
    client.AuthDelete(ctx, auth.UID)
    
    response := CheckEmailsResponse{
        Success:    true,
        Message:    fmt.Sprintf("Successfully processed %d messages", processedCount),
        EmailCount: processedCount,
        Errors:     errors,
    }
    
    json.NewEncoder(w).Encode(response)
}
