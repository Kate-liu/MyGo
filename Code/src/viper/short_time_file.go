package main

import (
	"github.com/spf13/viper"
)

func main() {
	v := viper.NewWithOptions(viper.KeyDelimiter("::"))
	v.SetDefault("chart::values", map[string]interface{}{
		"ingress": map[string]interface{}{
			"annotations": map[string]interface{}{
				"traefik.frontend.rule.type": "PathPrefix", "traefik.ingress.kubernetes.io/ssl-redirect": "true",
			},
		},
	})
	type config struct {
		Chart struct {
			Values map[string]interface{}
		}
	}

	var C config
	v.Unmarshal(&C)

}
