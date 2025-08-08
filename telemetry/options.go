package telemetry

import "log/slog"

// ------------------------------------
// --- Telemetry Config and Options ---
// ------------------------------------

// Option defines a function type for configuring telemetry options.
type Option func(*config)

// config holds the configuration for telemetry components like MySQL, NATS, Sentry, slog, and tracing.
type config struct {
	ServiceName   string
	MysqlConfig   mySQLConfig
	MysqlEnabled  bool
	NatsConfig    natsConfig
	NatsEnabled   bool
	SentryConfig  sentryConfig
	SentryEnabled bool
	SlogConfig    slogConfig
	SlogEnabled   bool
	TraceConfig   traceConfig
	TraceEnabled  bool
}

// -------------------------------
// --- Slog Config and Options ---
// -------------------------------

// WithSlog enables slog logging.
func WithSlog(opts ...SlogOption) Option {
	return func(cfg *config) {
		cfg.SlogEnabled = true
		slc := slogConfig{}
		for _, opt := range opts {
			opt(&slc)
		}
		cfg.SlogConfig = slc
	}
}

type slogConfig struct {
	logLevel slog.Level
	// Add more as needed
}

// SlogOption defines a function type for configuring slog options.
type SlogOption func(*slogConfig)

// SlogLogLevel sets the log level for slog.
func SlogLogLevel(level slog.Level) SlogOption {
	return func(cfg *slogConfig) { cfg.logLevel = level }
}

// ---------------------------------
// --- Sentry Config and Options ---
// ---------------------------------

// WithSentry enables Sentry error tracking.
func WithSentry(opts ...SentryOption) Option {
	return func(cfg *config) {
		cfg.SentryEnabled = true
		sc := sentryConfig{}
		for _, opt := range opts {
			opt(&sc)
		}
		cfg.SentryConfig = sc
	}
}

type sentryConfig struct {
	DSN         string
	Environment string
	Release     string
	// Add more as needed
}

// SentryOption defines a function type for configuring Sentry options.
type SentryOption func(*sentryConfig)

// SentryDSN sets the Data Source Name (DSN) for Sentry.
func SentryDSN(dsn string) SentryOption {
	return func(cfg *sentryConfig) { cfg.DSN = dsn }
}

// SentryEnvironment sets the environment for Sentry (e.g., "production", "development").
func SentryEnvironment(env string) SentryOption {
	return func(cfg *sentryConfig) { cfg.Environment = env }
}

// SentryRelease sets the release version for Sentry.
func SentryRelease(rel string) SentryOption {
	return func(cfg *sentryConfig) { cfg.Release = rel }
}

// -----------------------------------
// --- Traceing Config and Options ---
// -----------------------------------

// WithTrace enables tracing and allows configuration through options.
func WithTrace(opts ...TraceOption) Option {
	return func(cfg *config) {
		cfg.TraceEnabled = true
		tc := traceConfig{}
		for _, opt := range opts {
			opt(&tc)
		}
		cfg.TraceConfig = tc
	}
}

type traceConfig struct {
	ExporterURL string
	// Add more as needed
}

// TraceOption defines a function type for configuring trace options.
type TraceOption func(*traceConfig)

// TraceExporterURL sets the URL for the trace exporter.
func TraceExporterURL(url string) TraceOption {
	return func(cfg *traceConfig) { cfg.ExporterURL = url }
}

// --------------------------------
// --- MySQL Config and Options ---
// --------------------------------

type mySQLConfig struct {
	DSN string // Data Source Name for MySQL connection
	// Add more as needed
}

// MySQLOption defines a function type for configuring MySQL options.
type MySQLOption func(*mySQLConfig)

// WithMySQL enables MySQL database connection and allows configuration through options.
func WithMySQL(opts ...MySQLOption) Option {
	return func(cfg *config) {
		cfg.MysqlEnabled = true
		mc := mySQLConfig{}
		for _, opt := range opts {
			opt(&mc)
		}
		cfg.MysqlConfig = mc
	}
}

// MySQLDSN sets the Data Source Name for MySQL connection.
func MySQLDSN(dsn string) MySQLOption {
	return func(cfg *mySQLConfig) { cfg.DSN = dsn }
}

// -------------------------------
// --- NATS Config and Options ---
// -------------------------------

type natsConfig struct {
	URL string // NATS server URL
	// Add more as needed
}

// NATSOption defines a function type for configuring NATS options.
type NATSOption func(*natsConfig)

// WithNATS enables NATS messaging and allows configuration through options.
func WithNATS(opts ...NATSOption) Option {
	return func(cfg *config) {
		cfg.NatsEnabled = true
		nc := natsConfig{}
		for _, opt := range opts {
			opt(&nc)
		}
		cfg.NatsConfig = nc
	}
}

// NATSURL sets the NATS server URL.
func NATSURL(url string) NATSOption {
	return func(cfg *natsConfig) { cfg.URL = url }
}
