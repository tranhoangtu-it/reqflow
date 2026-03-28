package commands

import (
	"context"

	"github.com/ye-kart/reqflow/internal/domain"
	"github.com/ye-kart/reqflow/internal/platform/config"
)

type contextKey string

const (
	configKey     contextKey = "config"
	configOptsKey contextKey = "configOpts"
)

func withConfig(ctx context.Context, cfg *domain.AppConfig) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, configKey, cfg)
}

func configFromContext(ctx context.Context) *domain.AppConfig {
	if ctx == nil {
		return nil
	}
	cfg, _ := ctx.Value(configKey).(*domain.AppConfig)
	return cfg
}

func withConfigOpts(ctx context.Context, opts []config.LoadOption) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, configOptsKey, opts)
}

func configOptsFromContext(ctx context.Context) []config.LoadOption {
	if ctx == nil {
		return nil
	}
	opts, _ := ctx.Value(configOptsKey).([]config.LoadOption)
	return opts
}
