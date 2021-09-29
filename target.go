package consul

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-playground/form"
	"github.com/hashicorp/consul/api"
)

type target struct {
	// consul client params
	Addr        string        `form:"-"`
	User        string        `form:"-"`
	Password    string        `form:"-"`
	Token       string        `form:"token"`
	Wait        time.Duration `form:"wait"`
	Timeout     time.Duration `form:"timeout"`
	TLSInsecure bool          `form:"insecure"`

	// service query params
	Healthy           bool          `form:"healthy"`
	AllowStale        bool          `form:"allow-stale"`
	RequireConsistent bool          `form:"require-consistent"`
	Dc                string        `form:"dc"`
	Service           string        `form:"-"`
	Near              string        `form:"near"`
	MaxBackoff        time.Duration `form:"max-backoff"`
	Limit             int           `form:"limit"`

	Tag  string   `form:"tag"`
	tags []string `form:"-"`

	Sort string `form:"sort"`

	// TODO(mbobakov): custom parameters for the http-transport
	// TODO(mbobakov): custom parameters for the TLS subsystem
}

func (t *target) String() string {
	return fmt.Sprintf("service='%s' healthy='%t' tag='%s'", t.Service, t.Healthy, t.Tag)
}

var decoder = form.NewDecoder()

func init() {
	decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		return time.ParseDuration(vals[0])
	}, time.Duration(0))
}

// parseURL with parameters.
// see README.md for the actual format.
// URL schema will stay stable in the future for backward compatibility.
func parseURL(u string) (target, error) {
	rawURL, err := url.Parse(u)
	if err != nil {
		return target{}, fmt.Errorf("malformed URL: %w", err)
	}

	if rawURL.Scheme != schemeName ||
		len(rawURL.Host) == 0 || len(strings.TrimLeft(rawURL.Path, "/")) == 0 {
		return target{},
			fmt.Errorf("malformed URL('%s'). Must be in the next format: 'consul://[user:passwd]@host/service?param=value'", u)
	}

	tgt := target{
		User:    rawURL.User.Username(),
		Addr:    rawURL.Host,
		Service: strings.TrimLeft(rawURL.Path, "/"),
	}
	tgt.Password, _ = rawURL.User.Password()

	if err := decoder.Decode(&tgt, rawURL.Query()); err != nil {
		return target{}, fmt.Errorf("malformed URL parameters: %w", err)
	}

	if len(tgt.Near) == 0 {
		tgt.Near = "_agent"
	}

	if tgt.MaxBackoff == 0 {
		tgt.MaxBackoff = time.Second
	}

	if tgt.Tag != "" {
		tgt.tags = strings.Split(tgt.Tag, ",")
	}
	return tgt, nil
}

// consulConfig returns config based on the parsed target.
// It uses custom http-client.
func (t *target) consulConfig() *api.Config {
	var creds *api.HttpBasicAuth
	if len(t.User) > 0 && len(t.Password) > 0 {
		creds = new(api.HttpBasicAuth)
		creds.Password = t.Password
		creds.Username = t.User
	}

	return &api.Config{
		Address:  t.Addr,
		HttpAuth: creds,
		WaitTime: t.Wait,
		HttpClient: &http.Client{
			Timeout: t.Timeout,
		},
		TLSConfig: api.TLSConfig{
			InsecureSkipVerify: t.TLSInsecure,
		},
		Token: t.Token,
	}
}
