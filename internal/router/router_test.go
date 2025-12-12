package router

import (
	"testing"

	"github.com/amfib87/go-musthave-shortener-tpl/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		cfg    *config.Cnfg
		router bool
	}{
		{name: "Succs", cfg: &config.Cnfg{ServRunAddr: "", AddrForURL: ""}, router: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Init(tt.cfg)
			// TODO: update the condition below to compare got with tt.want.
			assert.NotNil(t, got, "Объект = nil")
		})
	}
}
