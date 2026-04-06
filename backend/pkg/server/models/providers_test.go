package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderTypeValid(t *testing.T) {
	t.Parallel()

	validTypes := []struct {
		name string
		pt   ProviderType
	}{
		{"openai", ProviderType("openai")},
		{"anthropic", ProviderType("anthropic")},
		{"gemini", ProviderType("gemini")},
		{"bedrock", ProviderType("bedrock")},
		{"ollama", ProviderType("ollama")},
		{"custom", ProviderType("custom")},
		{"deepseek", ProviderType("deepseek")},
		{"glm", ProviderType("glm")},
		{"kimi", ProviderType("kimi")},
		{"qwen", ProviderType("qwen")},
	}

	for _, tt := range validTypes {
		tt := tt
		t.Run("valid_"+tt.name, func(t *testing.T) {
			t.Parallel()
			assert.NoError(t, tt.pt.Valid())
		})
	}

	invalidTypes := []struct {
		name string
		pt   ProviderType
	}{
		{"empty", ProviderType("")},
		{"unknown", ProviderType("unknown")},
		{"gpt4", ProviderType("gpt4")},
		{"azure", ProviderType("azure")},
	}

	for _, tt := range invalidTypes {
		tt := tt
		t.Run("invalid_"+tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.pt.Valid()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid ProviderType")
		})
	}
}

func TestProviderTypeString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "openai", ProviderType("openai").String())
	assert.Equal(t, "anthropic", ProviderType("anthropic").String())
}

func TestProviderValid(t *testing.T) {
	t.Parallel()

	validProvider := Provider{
		UserID: 1,
		Type:   ProviderType("openai"),
		Name:   "my-openai",
		Config: json.RawMessage(`{"api_key":"test"}`),
	}

	t.Run("valid provider", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, validProvider.Valid())
	})

	t.Run("missing name", func(t *testing.T) {
		t.Parallel()
		p := validProvider
		p.Name = ""
		assert.Error(t, p.Valid())
	})

	t.Run("missing config", func(t *testing.T) {
		t.Parallel()
		p := validProvider
		p.Config = nil
		assert.Error(t, p.Valid())
	})

	t.Run("invalid type", func(t *testing.T) {
		t.Parallel()
		p := validProvider
		p.Type = ProviderType("invalid")
		assert.Error(t, p.Valid())
	})
}

func TestProviderTableName(t *testing.T) {
	t.Parallel()
	p := &Provider{}
	assert.Equal(t, "providers", p.TableName())
}

func TestCreateProviderValid(t *testing.T) {
	t.Parallel()

	t.Run("valid create provider", func(t *testing.T) {
		t.Parallel()
		cp := CreateProvider{Config: json.RawMessage(`{"key":"val"}`)}
		assert.NoError(t, cp.Valid())
	})

	t.Run("missing config", func(t *testing.T) {
		t.Parallel()
		cp := CreateProvider{Config: nil}
		assert.Error(t, cp.Valid())
	})
}

func TestPatchProviderValid(t *testing.T) {
	t.Parallel()

	t.Run("valid patch with name", func(t *testing.T) {
		t.Parallel()
		name := "updated"
		pp := PatchProvider{Name: &name}
		assert.NoError(t, pp.Valid())
	})

	t.Run("valid empty patch", func(t *testing.T) {
		t.Parallel()
		pp := PatchProvider{}
		assert.NoError(t, pp.Valid())
	})
}

func TestProviderInfoValid(t *testing.T) {
	t.Parallel()

	t.Run("valid provider info", func(t *testing.T) {
		t.Parallel()
		pi := ProviderInfo{Name: "my-provider", Type: ProviderType("openai")}
		assert.NoError(t, pi.Valid())
	})

	t.Run("missing name", func(t *testing.T) {
		t.Parallel()
		pi := ProviderInfo{Name: "", Type: ProviderType("openai")}
		assert.Error(t, pi.Valid())
	})

	t.Run("invalid type", func(t *testing.T) {
		t.Parallel()
		pi := ProviderInfo{Name: "my-provider", Type: ProviderType("invalid")}
		assert.Error(t, pi.Valid())
	})
}
