package envconf

const defaultDotEnvPath = ".env"

type loaderOptions struct {
	dotEnvEnabled bool
	dotEnvPath    string
}

type Option func(*loaderOptions)

func WithDotEnvPath(path string) Option {
	return func(o *loaderOptions) {
		o.dotEnvEnabled = true
		o.dotEnvPath = path
	}
}

func WithoutDotEnv() Option {
	return func(o *loaderOptions) {
		o.dotEnvEnabled = false
	}
}

func defaultOptions() loaderOptions {
	return loaderOptions{
		dotEnvEnabled: true,
		dotEnvPath:    defaultDotEnvPath,
	}
}

func fingerprint(opts loaderOptions) string {
	if !opts.dotEnvEnabled {
		return "dotenv=disabled"
	}
	return "dotenv=enabled;path=" + opts.dotEnvPath
}
