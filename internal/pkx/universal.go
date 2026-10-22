package pkx

import (
	"strings"
)

type universalFamily struct {
	id       string
	aliases  []string
	variants map[string]string
}

var curatedFamilies = []universalFamily{
	{"http-client", []string{"http-client", "axios", "httpx", "requests", "guzzle", "guzzlehttp/guzzle", "reqwest", "resty"}, map[string]string{"node": "axios", "python": "httpx", "php": "guzzlehttp/guzzle", "rust": "reqwest", "go": "github.com/go-resty/resty/v2"}},
	{"structured-logging", []string{"structured-logging", "pino", "structlog", "monolog", "monolog/monolog", "tracing", "zap"}, map[string]string{"node": "pino", "python": "structlog", "php": "monolog/monolog", "rust": "tracing", "go": "go.uber.org/zap"}},
	{"schema-validation", []string{"schema-validation", "zod", "pydantic", "symfony-validator", "symfony/validator", "validator"}, map[string]string{"node": "zod", "python": "pydantic", "php": "symfony/validator", "rust": "validator", "go": "github.com/go-playground/validator/v10"}},
}

func resolveCurated(query, ecosystem string) (string, string, bool) {
	for _, f := range curatedFamilies {
		for _, a := range f.aliases {
			if strings.EqualFold(query, a) {
				name, ok := f.variants[ecosystem]
				return name, f.id, ok
			}
		}
	}
	return query, "", true // unknown queries are treated as exact native package names
}
