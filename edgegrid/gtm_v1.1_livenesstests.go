package edgegrid

type GtmLivenessTest struct {
	Name                          string                     `json:"name"`
	TestInterval                  int                        `json:"testInterval"` // Minimum 10
	DisableNonstandardPortWarning bool                       `json:"disableNonstandardPortWarning"`
	HostHeader                    string                     `json:"hostHeader"`
	HTTPError3Xx                  bool                       `json:"httpError3xx"`
	HTTPError4Xx                  bool                       `json:"httpError4xx"`
	HTTPError5Xx                  bool                       `json:"httpError5xx"`
	RequestString                 string                     `json:"requestString"` // Required if testObjectProtocol is tcp or tcps
	ResponseString                string                     `json:"responseString"`
	SslClientCertificate          string                     `json:"sslClientCertificate"`
	SslClientPrivateKey           string                     `json:"sslClientPrivateKey"`
	TestObject                    string                     `json:"testObject"` // Required if testObjectProtocol is http or https.  Matches /^:[\d]+/.*$/
	TestObjectProtocol            GtmTestObjectProtocolValue `json:"testObjectProtocol"`
	TestObjectPort                int                        `json:"testObjectPort"`     // Range: 0 - 65535
	TestObjectUsername            string                     `json:"testObjectUsername"` // Required if testObjectProtocol is ftp
	TestObjectPassword            string                     `json:"testObjectPassword"` // Required if testObjectProtocol is ftp
	TestTimeout                   float64                    `json:"testTimeout"`        // Range: 0.001s - 60s
}

type GtmTestObjectProtocolValue string

const (
	GTM_TEST_OBJECT_PROTOCOL_HTTP  GtmTestObjectProtocolValue = "HTTP"
	GTM_TEST_OBJECT_PROTOCOL_HTTPS GtmTestObjectProtocolValue = "HTTPS"
	GTM_TEST_OBJECT_PROTOCOL_FTP   GtmTestObjectProtocolValue = "FTP"
	GTM_TEST_OBJECT_PROTOCOL_POP   GtmTestObjectProtocolValue = "POP"
	GTM_TEST_OBJECT_PROTOCOL_POPS  GtmTestObjectProtocolValue = "POPS"
	GTM_TEST_OBJECT_PROTOCOL_SMTP  GtmTestObjectProtocolValue = "SMTP"
	GTM_TEST_OBJECT_PROTOCOL_SMTPS GtmTestObjectProtocolValue = "SMTPS"
	GTM_TEST_OBJECT_PROTOCOL_TCP   GtmTestObjectProtocolValue = "TCP"
	GTM_TEST_OBJECT_PROTOCOL_TCPS  GtmTestObjectProtocolValue = "TCPS"
)
