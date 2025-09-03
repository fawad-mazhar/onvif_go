package utils

// GenerateGenericResponse generates a generic SOAP response
func GenerateGenericResponse() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
    <SOAP-ENV:Body>
        <GenericResponse/>
    </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`
}
