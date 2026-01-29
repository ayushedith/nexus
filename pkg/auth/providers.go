package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Provider interface {
	Apply(req *http.Request) error
}

type BearerAuth struct {
	Token string
}

func (a *BearerAuth) Apply(req *http.Request) error {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.Token))
	return nil
}

type BasicAuth struct {
	Username string
	Password string
}

func (a *BasicAuth) Apply(req *http.Request) error {
	req.SetBasicAuth(a.Username, a.Password)
	return nil
}

type APIKeyAuth struct {
	Key      string
	Value    string
	Location string
}

func (a *APIKeyAuth) Apply(req *http.Request) error {
	switch strings.ToLower(a.Location) {
	case "header":
		req.Header.Set(a.Key, a.Value)
	case "query":
		q := req.URL.Query()
		q.Set(a.Key, a.Value)
		req.URL.RawQuery = q.Encode()
	default:
		return fmt.Errorf("unsupported location: %s", a.Location)
	}
	return nil
}

type AWSSignatureV4 struct {
	AccessKey string
	SecretKey string
	Region    string
	Service   string
}

func (a *AWSSignatureV4) Apply(req *http.Request) error {
	now := time.Now().UTC()
	datestamp := now.Format("20060102")
	amzdate := now.Format("20060102T150405Z")

	req.Header.Set("X-Amz-Date", amzdate)

	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", datestamp, a.Region, a.Service)

	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-date:%s\n", req.Host, amzdate)
	signedHeaders := "host;x-amz-date"

	payloadHash := sha256Hash("")
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		req.Method,
		req.URL.Path,
		req.URL.RawQuery,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	)

	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s",
		amzdate,
		credentialScope,
		sha256Hash(canonicalRequest),
	)

	signingKey := a.getSignatureKey(a.SecretKey, datestamp, a.Region, a.Service)
	signature := hmacSHA256Hex(signingKey, stringToSign)

	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		a.AccessKey,
		credentialScope,
		signedHeaders,
		signature,
	)

	req.Header.Set("Authorization", authHeader)
	return nil
}

func (a *AWSSignatureV4) getSignatureKey(key, dateStamp, regionName, serviceName string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+key), dateStamp)
	kRegion := hmacSHA256(kDate, regionName)
	kService := hmacSHA256(kRegion, serviceName)
	kSigning := hmacSHA256(kService, "aws4_request")
	return kSigning
}

func sha256Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func hmacSHA256Hex(key []byte, data string) string {
	return fmt.Sprintf("%x", hmacSHA256(key, data))
}

func NewProvider(authType string, config map[string]string) (Provider, error) {
	switch strings.ToLower(authType) {
	case "bearer":
		return &BearerAuth{Token: config["token"]}, nil

	case "basic":
		return &BasicAuth{
			Username: config["username"],
			Password: config["password"],
		}, nil

	case "apikey":
		return &APIKeyAuth{
			Key:      config["key"],
			Value:    config["value"],
			Location: config["location"],
		}, nil

	case "aws":
		return &AWSSignatureV4{
			AccessKey: config["accessKey"],
			SecretKey: config["secretKey"],
			Region:    config["region"],
			Service:   config["service"],
		}, nil

	default:
		return nil, fmt.Errorf("unsupported auth type: %s", authType)
	}
}

func ApplyAuth(req *http.Request, authType string, config map[string]string) error {
	if authType == "" {
		return nil
	}

	provider, err := NewProvider(authType, config)
	if err != nil {
		return err
	}

	return provider.Apply(req)
}
