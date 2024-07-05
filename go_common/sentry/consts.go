package sentry

type ContextKey string

var (
	ContextHubKey  ContextKey = "sentry_hub"
	HookDriverName            = "mysql_with_sentry_wrapper"
)
