package dapr_test

import (
	"strings"
	"testing"

	dapr "github.com/pleme-io/dapr-component-generator"
	"go.yaml.in/yaml/v3"
)

func TestBackendComponentType(t *testing.T) {
	tests := []struct {
		name    string
		backend dapr.Backend
		want    string
	}{
		{"vault", dapr.BackendHashicorpVault, "secretstores.hashicorp.vault"},
		{"aws_sm", dapr.BackendAWSSecretManager, "secretstores.aws.secretmanager"},
		{"azure", dapr.BackendAzureKeyVault, "secretstores.azure.keyvault"},
		{"gcp", dapr.BackendGCPSecretManager, "secretstores.gcp.secretmanager"},
		{"local_file", dapr.BackendLocalFile, "secretstores.local.file"},
		{"k8s", dapr.BackendKubernetes, "secretstores.kubernetes"},
		{"custom", dapr.Backend("custom.backend"), "secretstores.custom.backend"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := dapr.New("x", tt.backend)
			out := string(ss.MustRender())
			if !strings.Contains(out, "type: "+tt.want) {
				t.Errorf("rendered type = %q, want type: %q\n%s", out, tt.want, out)
			}
		})
	}
}

func TestRenderContents(t *testing.T) {
	tests := []struct {
		name     string
		store    dapr.SecretStore
		contains []string
		absent   []string
	}{
		{
			name: "literal_metadata_and_scopes",
			store: dapr.New("vault", dapr.BackendHashicorpVault,
				dapr.WithNamespace("prod"),
				dapr.WithMetadata("vaultAddr", "https://vault:8200"),
				dapr.WithScope("orders"),
				dapr.WithScope("billing"),
			),
			contains: []string{
				"apiVersion: dapr.io/v1alpha1",
				"kind: Component",
				"name: vault",
				"namespace: prod",
				"type: secretstores.hashicorp.vault",
				"version: v1",
				"name: vaultAddr",
				"value: https://vault:8200",
				"scopes:",
				"- orders",
				"- billing",
			},
		},
		{
			name: "secret_key_ref",
			store: dapr.New("aws", dapr.BackendAWSSecretManager,
				dapr.WithSecretRef("accessKey", "aws-creds", "access-key-id"),
			),
			contains: []string{
				"secretKeyRef:",
				"name: aws-creds",
				"key: access-key-id",
			},
			absent: []string{"value:"},
		},
		{
			name:     "version_override",
			store:    dapr.New("x", dapr.BackendLocalEnv, dapr.WithVersion("v2")),
			contains: []string{"version: v2"},
		},
		{
			name:     "namespace_omitted_when_empty",
			store:    dapr.New("x", dapr.BackendLocalEnv),
			contains: []string{"name: x"},
			absent:   []string{"namespace:"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := string(tt.store.MustRender())
			for _, w := range tt.contains {
				if !strings.Contains(out, w) {
					t.Errorf("missing %q\n--- got ---\n%s", w, out)
				}
			}
			for _, a := range tt.absent {
				if strings.Contains(out, a) {
					t.Errorf("unexpectedly present %q\n--- got ---\n%s", a, out)
				}
			}
		})
	}
}

func TestRenderProducesValidYAML(t *testing.T) {
	ss := dapr.New("vault", dapr.BackendHashicorpVault,
		dapr.WithNamespace("prod"),
		dapr.WithMetadata("vaultAddr", "https://vault:8200"),
		dapr.WithSecretRef("token", "vault-token", "token"),
		dapr.WithScope("orders"),
	)
	out, err := dapr.Render(ss)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	// Round-trip: the rendered bytes must parse back into a generic map, proving
	// we emitted valid YAML (not raw text that merely looks right).
	var back map[string]any
	if err := yaml.Unmarshal(out, &back); err != nil {
		t.Fatalf("rendered output is not valid YAML: %v\n%s", err, out)
	}
	if back["apiVersion"] != "dapr.io/v1alpha1" {
		t.Errorf("apiVersion = %v, want dapr.io/v1alpha1", back["apiVersion"])
	}
	spec, ok := back["spec"].(map[string]any)
	if !ok {
		t.Fatalf("spec is not a map: %T", back["spec"])
	}
	if spec["type"] != "secretstores.hashicorp.vault" {
		t.Errorf("spec.type = %v", spec["type"])
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		store   dapr.SecretStore
		wantErr bool
	}{
		{"ok", dapr.New("x", dapr.BackendLocalEnv), false},
		{"missing_name", dapr.SecretStore{Backend: dapr.BackendLocalEnv}, true},
		{"missing_backend", dapr.SecretStore{Name: "x"}, true},
		{
			"metadata_missing_name",
			dapr.SecretStore{Name: "x", Backend: dapr.BackendLocalEnv, Metadata: []dapr.MetadataEntry{{Value: "v"}}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.store.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() err = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if _, rerr := tt.store.RenderYAML(); rerr == nil {
					t.Error("RenderYAML should fail on invalid store")
				}
			}
		})
	}
}

func TestFromConfig(t *testing.T) {
	cfg := dapr.Config{
		Name:      "aws-store",
		Namespace: "default",
		Backend:   "aws.secretmanager",
		Version:   "v1",
		Metadata: []dapr.MetadataField{
			{Name: "region", Value: "us-east-1"},
			{Name: "accessKey", SecretName: "aws-creds", SecretKey: "access-key-id"},
		},
		Scopes: []string{"orders", "billing"},
	}
	out := string(dapr.FromConfig(cfg).MustRender())
	wants := []string{
		"name: aws-store",
		"namespace: default",
		"type: secretstores.aws.secretmanager",
		"name: region",
		"value: us-east-1",
		"secretKeyRef:",
		"name: aws-creds",
		"key: access-key-id",
		"- orders",
		"- billing",
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("FromConfig missing %q\n--- got ---\n%s", w, out)
		}
	}
}

// TestFromConfigParsesFromYAMLManifest proves the shikumi-shaped Config struct
// loads from a real YAML manifest and produces the same component — the
// generator's end-to-end path (manifest in, metadata.yaml out).
func TestFromConfigParsesFromYAMLManifest(t *testing.T) {
	manifest := `
name: vault-store
backend: hashicorp.vault
namespace: prod
metadata:
  - name: vaultAddr
    value: https://vault:8200
  - name: token
    secretName: vault-token
    secretKey: token
scopes:
  - orders
`
	var cfg dapr.Config
	if err := yaml.Unmarshal([]byte(manifest), &cfg); err != nil {
		t.Fatalf("manifest did not parse: %v", err)
	}
	out := string(dapr.FromConfig(cfg).MustRender())
	for _, w := range []string{
		"name: vault-store",
		"type: secretstores.hashicorp.vault",
		"secretKeyRef:",
		"name: vault-token",
	} {
		if !strings.Contains(out, w) {
			t.Errorf("missing %q\n--- got ---\n%s", w, out)
		}
	}
}

func TestRenderableCarrier(t *testing.T) {
	var _ dapr.Renderable = dapr.SecretStore{Name: "x", Backend: dapr.BackendLocalEnv}
	out, err := dapr.Render(dapr.New("x", dapr.BackendLocalEnv))
	if err != nil || len(out) == 0 {
		t.Fatalf("Render via carrier failed: err=%v len=%d", err, len(out))
	}
}
