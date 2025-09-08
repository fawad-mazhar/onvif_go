// Package server provides ONVIF server implementations.
package server

import (
	"fmt"
	"net/http"
	"strings"

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
	logger.Infof("Starting notification server on port %d", cfg.NotificationPort)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		logger.Fatalf("Failed to start notification server: %v", err)
	}
	return nil
}

// handleNotificationRequest handles notification requests and sends appropriate responses
func handleNotificationRequest(w http.ResponseWriter, r *http.Request, cfg *config.ServiceContext) {
	// Read request using shared utility
	request, err := utils.ReadSOAPRequest(r, "notification")
	if err != nil {
		logger.Warnf("%v", err)
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}
	logger.Debugf("Notification request: %s", request)

	// Set response headers using shared utility
	utils.SetSOAPHeaders(w)

	// Handle different notification actions
	switch {
	case strings.Contains(request, "Subscribe"):
		// Handle subscribe request
		response := generateSubscribeResponse()
		if _, err := fmt.Fprint(w, response); err != nil {
			logger.Errorf("Failed to write subscribe response: %v", err)
		}
	case strings.Contains(request, "Unsubscribe"):
		// Handle unsubscribe request
		response := generateUnsubscribeResponse()
		if _, err := fmt.Fprint(w, response); err != nil {
			logger.Errorf("Failed to write unsubscribe response: %v", err)
		}
	case strings.Contains(request, "Renew"):
		// Handle renew request
		response := generateRenewResponse()
		if _, err := fmt.Fprint(w, response); err != nil {
			logger.Errorf("Failed to write renew response: %v", err)
		}
	case strings.Contains(request, "PullMessages"):
		// Handle pull messages request
		response := generatePullMessagesResponse(cfg)
		if _, err := fmt.Fprint(w, response); err != nil {
			logger.Errorf("Failed to write pull messages response: %v", err)
		}
	default:
		// Send a generic response
		response := utils.GenerateGenericResponse()
		if _, err := fmt.Fprint(w, response); err != nil {
			logger.Errorf("Failed to write generic response: %v", err)
		}
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

	// Replace placeholders with actual values using shared utility
	replacements := map[string]string{
		"%CURRENT_TIME%":     utils.GetCurrentTimeRFC3339(),
		"%TERMINATION_TIME%": utils.GetTimeAfterHoursRFC3339(1),
	}
	response := utils.ReplaceTemplatePlaceholders(template, replacements)

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

	// Replace placeholder with actual value using shared utility
	replacements := map[string]string{
		"%TERMINATION_TIME%": utils.GetTimeAfterHoursRFC3339(1),
	}
	response := utils.ReplaceTemplatePlaceholders(template, replacements)

	return response
}

// generatePullMessagesResponse generates a notification pull messages response
func generatePullMessagesResponse(_ *config.ServiceContext) string {
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

	// Replace placeholders with actual values using shared utility
	replacements := map[string]string{
		"%CURRENT_TIME%":     utils.GetCurrentTimeRFC3339(),
		"%TERMINATION_TIME%": utils.GetTimeAfterHoursRFC3339(1),
	}
	response := utils.ReplaceTemplatePlaceholders(template, replacements)

	return response
}
