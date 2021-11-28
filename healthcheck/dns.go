package healthcheck

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net"
)

// DNSHealthcheckConfiguration defines a DNS healthcheck configuration
type DNSHealthcheckConfiguration struct {
	Base        `json:",inline" yaml:",inline"`
	ExpectedIPs []IP   `json:"expected-ips,omitempty" yaml:"expected-ips,omitempty"`
	Domain      string `json:"domain"`
}

// DNSHealthcheck defines an HTTP healthcheck
type DNSHealthcheck struct {
	Logger *zap.Logger
	Config *DNSHealthcheckConfiguration
	URL    string

	Tick *time.Ticker
}

// Validate validates the healthcheck configuration
func (config *DNSHealthcheckConfiguration) Validate() error {
	if config.Base.Name == "" {
		return errors.New("The healthcheck name is missing")
	}
	if config.Domain == "" {
		return errors.New("The healthcheck domain is missing")
	}
	if !config.Base.OneOff {
		if config.Base.Interval < Duration(2*time.Second) {
			return errors.New("The healthcheck interval should be greater than 2 second")
		}
	}
	return nil
}

// Initialize the healthcheck.
func (h *DNSHealthcheck) Initialize() error {
	return nil
}

// GetConfig get the config
func (h *DNSHealthcheck) GetConfig() interface{} {
	return h.Config
}

// Base get the base configuration
func (h *DNSHealthcheck) Base() Base {
	return h.Config.Base
}

// SetSource set the healthcheck source
func (h *DNSHealthcheck) SetSource(source string) {
	h.Config.Base.Source = source
}

// Summary returns an healthcheck summary
func (h *DNSHealthcheck) Summary() string {
	summary := ""
	if h.Config.Base.Description != "" {
		summary = fmt.Sprintf("%s on %s", h.Config.Base.Description, h.Config.Domain)

	} else {
		summary = fmt.Sprintf("on %s", h.Config.Domain)
	}

	return summary
}

// LogError logs an error with context
func (h *DNSHealthcheck) LogError(err error, message string) {
	h.Logger.Error(err.Error(),
		zap.String("extra", message),
		zap.String("domain", h.Config.Domain),
		zap.String("name", h.Config.Base.Name))
}

// LogDebug logs a message with context
func (h *DNSHealthcheck) LogDebug(message string) {
	h.Logger.Debug(message,
		zap.String("domain", h.Config.Domain),
		zap.String("name", h.Config.Base.Name))
}

// LogInfo logs a message with context
func (h *DNSHealthcheck) LogInfo(message string) {
	h.Logger.Info(message,
		zap.String("domain", h.Config.Domain),
		zap.String("name", h.Config.Base.Name))
}

func verifyIPs(expectedIPs []IP, lookupIPs []net.IP) error {
	notFound := []string{}
	for i := range expectedIPs {
		netIP := net.IP(expectedIPs[i])
		found := false
		for j := range lookupIPs {
			respIP := lookupIPs[j]
			if netIP.Equal(respIP) {
				found = true
				break
			}
		}
		if !found {
			notFound = append(notFound, netIP.String())
		}
	}
	if len(notFound) != 0 {
		l := ""
		for _, notFound := range notFound {
			l = l + "," + notFound
		}
		return fmt.Errorf("Expected IP address not found. IPs found are %s", l)
	}
	return nil
}

// Execute executes an healthcheck on the given domain
func (h *DNSHealthcheck) Execute() error {
	h.LogDebug("start executing healthcheck")
	ips, err := net.LookupIP(h.Config.Domain)
	if err != nil {
		return errors.Wrapf(err, "Fail to lookup IP for domain")
	}
	err = verifyIPs(h.Config.ExpectedIPs, ips)
	if err != nil {
		return err
	}
	return nil
}

// NewDNSHealthcheck creates a DNS healthcheck from a logger and a configuration
func NewDNSHealthcheck(logger *zap.Logger, config *DNSHealthcheckConfiguration) *DNSHealthcheck {
	return &DNSHealthcheck{
		Logger: logger,
		Config: config,
	}
}

// MarshalJSON marshal to json a dns healthcheck
func (h *DNSHealthcheck) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.Config)
}
