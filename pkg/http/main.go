package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Haameed/f5_bigip_exporter/internal/config"
	"github.com/Haameed/f5_bigip_exporter/internal/utils"
)

type BigIPHTTP interface {
	Get(path string, obj any) error
}

func NewBigIPClient(ctx context.Context, tgt url.URL, hc *http.Client, aConfig config.BigIpExporterConfig) (BigIPHTTP, error) {
	auth, ok := aConfig.AuthKeys[config.Target(tgt.String())]
	if !ok {
		return nil, fmt.Errorf("no API authentication registered for %q", tgt.String())
	}
	token, err := utils.GetToken(tgt.String(), auth.UserName, auth.Password, aConfig.TLSInsecure, time.Duration(aConfig.ScrapeTimeout))
	if err != nil {
		return nil, fmt.Errorf("failed to obtain token: %w", err)
	}
	if token.Token != "" {
		if tgt.Scheme != "https" {
			return nil, fmt.Errorf("we are Using token, So please use HTTPS scheme for %q", tgt.String())
		}
		c, err := newBigIPTokenClient(ctx, tgt, hc, token)
		if err != nil {
			return nil, err
		}
		return c, nil
	}
	return nil, fmt.Errorf("invalid authentication data for %q", tgt.String())
}

func Configure(config config.BigIpExporterConfig) error {

	tc := &tls.Config{}
	if config.TLSInsecure {
		tc.InsecureSkipVerify = true
	}
	http.DefaultTransport.(*http.Transport).TLSHandshakeTimeout = time.Duration(config.TLSTimeout) * time.Second
	http.DefaultTransport.(*http.Transport).TLSClientConfig = tc
	return nil
}
