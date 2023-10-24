package httpauthd

type ServerOption interface {
	apply(*serverOptions) error
}

type serverOptions struct {
}

func defaultServerOptions() serverOptions {
	return serverOptions{}
}

type funcServerOption struct {
	f func(*serverOptions) error
}

func (fdo *funcServerOption) apply(do *serverOptions) error {
	return fdo.f(do)
}

func newFuncServerOption(f func(*serverOptions) error) *funcServerOption {
	return &funcServerOption{
		f: f,
	}
}
