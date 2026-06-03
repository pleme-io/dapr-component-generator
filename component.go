package dapr

import (
	"fmt"

	"go.yaml.in/yaml/v3"
)

// Canonical Dapr Component apiVersion / kind / component version.
const (
	apiVersion         = "dapr.io/v1alpha1"
	kind               = "Component"
	defaultCompVersion = "v1"
)

// SecretStore is a typed Dapr SecretStore component. It renders to the canonical
// Dapr Component `metadata.yaml` via [SecretStore.RenderYAML]. The required
// pieces — Name and Backend — are set by [New]; everything else is optional.
type SecretStore struct {
	// Name is the component name (metadata.name).
	Name string
	// Namespace, when non-empty, is emitted as metadata.namespace.
	Namespace string
	// Backend selects the secret-store backend (spec.type = secretstores.<backend>).
	Backend Backend
	// Version is the component spec version; defaults to "v1" when empty.
	Version string
	// Metadata is the ordered list of spec.metadata entries (backend config).
	Metadata []MetadataEntry
	// Scopes restricts the component to the listed Dapr app IDs (scopes).
	Scopes []string
}

// MetadataEntry is one spec.metadata item: a name with either a literal value or
// a secretKeyRef. Exactly one of Value / SecretKeyRef is emitted (Value wins
// when both are set).
type MetadataEntry struct {
	Name         string
	Value        string
	SecretKeyRef *SecretKeyRef
}

// SecretKeyRef references a secret stored in another secret store
// (spec.metadata[].secretKeyRef).
type SecretKeyRef struct {
	Name string
	Key  string
}

// ── YAML wire structs (the emitter owns the shape) ─────────────
// These mirror the Dapr Component CRD exactly. Marshaling through them — rather
// than formatting strings — guarantees valid YAML: the emitter handles quoting,
// indentation, and special characters.

type componentDoc struct {
	APIVersion string      `yaml:"apiVersion"`
	Kind       string      `yaml:"kind"`
	Metadata   docMetadata `yaml:"metadata"`
	Spec       docSpec     `yaml:"spec"`
	Scopes     []string    `yaml:"scopes,omitempty"`
}

type docMetadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace,omitempty"`
}

type docSpec struct {
	Type     string         `yaml:"type"`
	Version  string         `yaml:"version"`
	Metadata []docMetaEntry `yaml:"metadata,omitempty"`
}

type docMetaEntry struct {
	Name         string        `yaml:"name"`
	Value        string        `yaml:"value,omitempty"`
	SecretKeyRef *docSecretRef `yaml:"secretKeyRef,omitempty"`
}

type docSecretRef struct {
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

// toDoc lowers the public SecretStore to the wire struct, applying defaults.
func (s SecretStore) toDoc() componentDoc {
	version := s.Version
	if version == "" {
		version = defaultCompVersion
	}
	entries := make([]docMetaEntry, 0, len(s.Metadata))
	for _, m := range s.Metadata {
		e := docMetaEntry{Name: m.Name}
		switch {
		case m.Value != "":
			e.Value = m.Value
		case m.SecretKeyRef != nil:
			e.SecretKeyRef = &docSecretRef{Name: m.SecretKeyRef.Name, Key: m.SecretKeyRef.Key}
		}
		entries = append(entries, e)
	}
	return componentDoc{
		APIVersion: apiVersion,
		Kind:       kind,
		Metadata:   docMetadata{Name: s.Name, Namespace: s.Namespace},
		Spec: docSpec{
			Type:     s.Backend.componentType(),
			Version:  version,
			Metadata: entries,
		},
		Scopes: s.Scopes,
	}
}

// Validate reports the first structural problem with the SecretStore, or nil.
func (s SecretStore) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("dapr: SecretStore name is required")
	}
	if s.Backend == "" {
		return fmt.Errorf("dapr: SecretStore %q backend is required", s.Name)
	}
	for i, m := range s.Metadata {
		if m.Name == "" {
			return fmt.Errorf("dapr: SecretStore %q metadata[%d] name is required", s.Name, i)
		}
	}
	return nil
}

// RenderYAML renders the component to canonical Dapr `metadata.yaml` bytes via
// the YAML emitter (go.yaml.in/yaml/v3). It is the carrier behind [Render]
// (borealis Law 5). It validates first; an invalid component returns an error
// rather than emitting a malformed manifest.
func (s SecretStore) RenderYAML() ([]byte, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return yaml.Marshal(s.toDoc())
}

// MustRender renders the component or panics — convenience for tests and
// generators where a malformed component is a programmer error.
func (s SecretStore) MustRender() []byte {
	out, err := s.RenderYAML()
	if err != nil {
		panic(err)
	}
	return out
}
