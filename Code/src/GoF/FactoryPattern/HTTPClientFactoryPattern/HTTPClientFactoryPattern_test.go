package HTTPClientFactoryPattern

import (
	"net/http"
	"testing"
)

// QueryUser testing
func QueryUser(doer Doer) error {
	req, err := http.NewRequest("Get", "https://iam.api.marmotedu.com:8080/v1/secrets", nil)
	if err != nil {
		return err
	}
	_, err = doer.Do(req)
	if err != nil {
		return err
	}
	return nil
}

func TestQueryUser(t *testing.T) {
	doer := NewMockHTTPClient()
	if err := QueryUser(doer); err != nil {
		t.Errorf("QueryUser failed, err: %v", err)
	}
}
