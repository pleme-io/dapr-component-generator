package dapr

// Option configures a [SecretStore] built by [New]. It follows the canonical
// fleet functional-options shape (the exported Option func type, GSDS §3.5 — the
// same shape as errors-go's errors.Option).
type Option func(*SecretStore)

// New builds a [SecretStore] from the required name and backend plus functional
// options. It is the canonical constructor shape New(required..., opts ...Option)
// — the required pieces a SecretStore cannot omit, then options.
func New(name string, backend Backend, opts ...Option) SecretStore {
	s := SecretStore{Name: name, Backend: backend}
	for _, opt := range opts {
		opt(&s)
	}
	return s
}

// WithNamespace sets the component namespace (metadata.namespace).
func WithNamespace(ns string) Option {
	return func(s *SecretStore) { s.Namespace = ns }
}

// WithVersion overrides the component spec version (default "v1").
func WithVersion(v string) Option {
	return func(s *SecretStore) { s.Version = v }
}

// WithMetadata appends a literal-valued spec.metadata entry.
func WithMetadata(name, value string) Option {
	return func(s *SecretStore) {
		s.Metadata = append(s.Metadata, MetadataEntry{Name: name, Value: value})
	}
}

// WithSecretRef appends a spec.metadata entry whose value is a secretKeyRef.
func WithSecretRef(name, refName, refKey string) Option {
	return func(s *SecretStore) {
		s.Metadata = append(s.Metadata, MetadataEntry{
			Name:         name,
			SecretKeyRef: &SecretKeyRef{Name: refName, Key: refKey},
		})
	}
}

// WithScope appends a Dapr app ID to the component scopes.
func WithScope(appID string) Option {
	return func(s *SecretStore) { s.Scopes = append(s.Scopes, appID) }
}

// WithScopes appends several Dapr app IDs to the component scopes.
func WithScopes(appIDs ...string) Option {
	return func(s *SecretStore) { s.Scopes = append(s.Scopes, appIDs...) }
}

// Config is the typed, declarative description of a SecretStore — the shikumi-
// shaped surface (a plain struct with yaml tags) a generator loads from a
// manifest. [FromConfig] turns it into the same [SecretStore] the imperative
// [New] builder produces.
type Config struct {
	Name      string          `yaml:"name"`
	Namespace string          `yaml:"namespace"`
	Backend   string          `yaml:"backend"`
	Version   string          `yaml:"version"`
	Metadata  []MetadataField `yaml:"metadata"`
	Scopes    []string        `yaml:"scopes"`
}

// MetadataField is the declarative form of a [MetadataEntry]. A non-empty
// SecretName promotes the entry to a secretKeyRef.
type MetadataField struct {
	Name       string `yaml:"name"`
	Value      string `yaml:"value"`
	SecretName string `yaml:"secretName"`
	SecretKey  string `yaml:"secretKey"`
}

// FromConfig builds a [SecretStore] from a typed [Config]. It is the
// config-driven peer of [New] (GSDS §3.5 — FromConfig(cfg)); a generator that
// loads a YAML manifest via shikumi gets the same [SecretStore] surface.
func FromConfig(cfg Config) SecretStore {
	opts := []Option{}
	if cfg.Namespace != "" {
		opts = append(opts, WithNamespace(cfg.Namespace))
	}
	if cfg.Version != "" {
		opts = append(opts, WithVersion(cfg.Version))
	}
	for _, m := range cfg.Metadata {
		if m.SecretName != "" {
			opts = append(opts, WithSecretRef(m.Name, m.SecretName, m.SecretKey))
		} else {
			opts = append(opts, WithMetadata(m.Name, m.Value))
		}
	}
	if len(cfg.Scopes) > 0 {
		opts = append(opts, WithScopes(cfg.Scopes...))
	}
	return New(cfg.Name, Backend(cfg.Backend), opts...)
}
