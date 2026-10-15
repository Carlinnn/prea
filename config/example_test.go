package config_test

import (
	"fmt"
	"time"

	"github.com/prea-framework/prea/config"
)

// Value decodes a key into any type; a missing key leaves the value as it is.
func ExampleConfig_Value() {
	cfg := config.New("dev", map[string]any{
		"mail": map[string]any{"addr": "smtp.example.com:587", "timeout": "30s"},
	})

	var addr string
	var timeout time.Duration
	retries := 3
	_ = cfg.Value("mail.addr", &addr)
	_ = cfg.Value("mail.timeout", &timeout)
	_ = cfg.Value("mail.retries", &retries) // not in the config: stays 3

	fmt.Println(cfg.Profile(), addr, timeout, retries)
	// Output: dev smtp.example.com:587 30s 3
}
