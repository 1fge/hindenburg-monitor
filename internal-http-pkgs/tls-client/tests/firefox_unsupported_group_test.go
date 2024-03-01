package tests

import (
	"strings"
	"testing"

	http "github.com/1fge/hindenburg-monitor/internal-http-pkgs/fhttp"
	tls_client "github.com/1fge/hindenburg-monitor/internal-http-pkgs/tls-client"
	"github.com/1fge/hindenburg-monitor/internal-http-pkgs/tls-client/profiles"
	"github.com/stretchr/testify/assert"
)

func TestWeb(t *testing.T) {
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Firefox_110),
	}

	client, err := tls_client.NewHttpClient(nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://registrierung.web.de/account/email-registration", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}

	req.Header = defaultHeader

	_, err = client.Do(req)
	assert.NoError(t, err)
}
