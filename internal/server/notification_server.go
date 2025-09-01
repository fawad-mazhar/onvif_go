package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	
	"github.com/fawad-mazhar/onvif-go/internal/config"
	"github.com/fawad-mazhar/onvif-go/internal/logger"
	"github.com/fawad-mazhar/onvif-go/internal/utils"
)

func StartNotificationServer(cfg *config.ServiceContext) error {
	// Initialize logging
	logger.InitLogger(logger.INFO)
	// logger.SetLevel(logger.INFO) - level already set during initialization
	
	
	// Create HTTP server for notifications
	http.HandleFunc("/notification", func(w http.ResponseWriter, r *http.Request) {
		// Handle notification requests
		handleNotificationRequest(w, r, cfg)
	})
	
	// Start the HTTP server
	addr := fmt.Sprintf(":%d", cfg.NotificationPort)
	logger.Info("Starting notification server on port %d", cfg.NotificationPort)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		logger.Fatal("Failed to start notification server: %v", err)
	}
	return nil
}

// handleNotificationRequest handles notification requests and sends appropriate responses
func handleNotificationRequest(w http.ResponseWriter, r *http.Request, cfg *config.ServiceContext) {
	// Read the request body
	body := make([]byte, 1024)
	n, err := r.Body.Read(body)
	if err != nil {
		logger.Warn("Failed to read notification request: %v", err)
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}
	
	request := string(body[:n])
	logger.Debug("Notification request: %s", request)
	
	// Set response headers
	w.Header().Set("Content-Type", "application/soap+xml")
	
	// Handle different notification actions
	switch {
	case strings.Contains(request, "Subscribe"):
		// Handle subscribe request
		response := generateSubscribeResponse()
		fmt.Fprint(w, response)
	case strings.Contains(request, "Unsubscribe"):
		// Handle unsubscribe request
		response := generateUnsubscribeResponse()
		fmt.Fprint(w, response)
	case strings.Contains(request, "Renew"):
		// Handle renew request
		response := generateRenewResponse()
		fmt.Fprint(w, response)
	case strings.Contains(request, "PullMessages"):
		// Handle pull messages request
		response := generatePullMessagesResponse(cfg)
		fmt.Fprint(w, response)
	default:
		// Send a generic response
		response := utils.GenerateGenericResponse()
		fmt.Fprint(w, response)
	}
}

// generateSubscribeResponse generates a notification subscribe response
func generateSubscribeResponse() string {
	// Read the Subscribe template
	template := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope"
                   xmlns:SOAP-ENC="http://www.w3.org/2003/05/soap-encoding"
                   xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
                   xmlns:xsd="http://www.w3.org/2001/XMLSchema"
                   xmlns:wsa="http://www.w3.org/2005/08/addressing"
                   xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2">
    <SOAP-ENV:Header/>
    <SOAP-ENV:Body>
        <wsnt:SubscribeResponse>
            <wsnt:SubscriptionReference>
                <wsa:Address>http://example.com/notifications</wsa:Address>
            </wsnt:SubscriptionReference>
            <wsnt:CurrentTime>%CURRENT_TIME%</wsnt:CurrentTime>
            <wsnt:TerminationTime>%TERMINATION_TIME%</wsnt:TerminationTime>
        </wsnt:SubscribeResponse>
    </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`
	
	// Replace placeholders with actual values
	currentTime := time.Now().Format(time.RFC3339)
	terminationTime := time.Now().Add(time.Hour).Format(time.RFC3339) // 1 hour from now
	
	response := strings.Replace(template, "%CURRENT_TIME%", currentTime, -1)
	response = strings.Replace(response, "%TERMINATION_TIME%", terminationTime, -1)
	
	return response
}

// generateUnsubscribeResponse generates a notification unsubscribe response
func generateUnsubscribeResponse() string {
	// Read the Unsubscribe template
	template := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope"
                   xmlns:SOAP-ENC="http://www.w3.org/2003/05/soap-encoding"
                   xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
                   xmlns:xsd="http://www.w3.org/2001/XMLSchema"
                   xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2">
    <SOAP-ENV:Header/>
    <SOAP-ENV:Body>
        <wsnt:UnsubscribeResponse/>
    </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`
	
	return template
}

// generateRenewResponse generates a notification renew response
func generateRenewResponse() string {
	// Read the Renew template
	template := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope"
                   xmlns:SOAP-ENC="http://www.w3.org/2003/05/soap-encoding"
                   xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
                   xmlns:xsd="http://www.w3.org/2001/XMLSchema"
                   xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2">
    <SOAP-ENV:Header/>
    <SOAP-ENV:Body>
        <wsnt:RenewResponse>
            <wsnt:TerminationTime>%TERMINATION_TIME%</wsnt:TerminationTime>
        </wsnt:RenewResponse>
    </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`
	
	// Replace placeholder with actual value
	terminationTime := time.Now().Add(time.Hour).Format(time.RFC3339) // 1 hour from now
	
	response := strings.Replace(template, "%TERMINATION_TIME%", terminationTime, -1)
	
	return response
}

// generatePullMessagesResponse generates a notification pull messages response
func generatePullMessagesResponse(cfg *config.ServiceContext) string {
	// Read the PullMessages template
	template := `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope"
                   xmlns:SOAP-ENC="http://www.w3.org/2003/05/soap-encoding"
                   xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
                   xmlns:xsd="http://www.w3.org/2001/XMLSchema"
                   xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2"
                   xmlns:tt="http://www.onvif.org/ver10/schema">
    <SOAP-ENV:Header/>
    <SOAP-ENV:Body>
        <wsnt:PullMessagesResponse>
            <wsnt:CurrentTime>%CURRENT_TIME%</wsnt:CurrentTime>
            <wsnt:TerminationTime>%TERMINATION_TIME%</wsnt:TerminationTime>
        </wsnt:PullMessagesResponse>
    </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`
	
	// Replace placeholders with actual values
	currentTime := time.Now().Format(time.RFC3339)
	terminationTime := time.Now().Add(time.Hour).Format(time.RFC3339) // 1 hour from now
	
	response := strings.Replace(template, "%CURRENT_TIME%", currentTime, -1)
	response = strings.Replace(response, "%TERMINATION_TIME%", terminationTime, -1)
	
	return response
}

