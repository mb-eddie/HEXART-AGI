package agent

import (
	"testing"

	"github.com/Autumn-27/norma/llm"
)

// ARTEX_LLM_PROVIDER is matched regardless of case and surrounding spaces. A
// miss used to fall through to anthropic: "Ollama" then sent ANTHROPIC_API_KEY
// to the Ollama base URL.
func TestFromEnvProviderIsCaseInsensitive(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-secret")
	t.Setenv("OPENAI_API_KEY", "sk-oai")
	t.Setenv("NVIDIA_API_KEY", "nvapi-test")
	t.Setenv("ARTEX_LLM_BASE_URL", "")
	t.Setenv("ARTEX_LLM_MODEL", "")
	for _, tc := range []struct {
		env  string
		want llm.Format
		key  string
	}{
		{"Ollama", FormatOllama, ""},
		{" OLLAMA ", FormatOllama, ""},
		{"OpenAI", llm.FormatOpenAI, "sk-oai"},
		{"OpenAI-Responses", llm.FormatOpenAIResponses, "sk-oai"},
		{"Nemotron", llm.FormatOpenAI, "nvapi-test"},
		{"Anthropic", llm.FormatAnthropic, "sk-ant-secret"},
	} {
		t.Setenv("ARTEX_LLM_PROVIDER", tc.env)
		c, ok := FromEnv()
		if !ok {
			t.Fatalf("%q: FromEnv returned ok=false", tc.env)
		}
		if c.Format != tc.want || c.APIKey != tc.key {
			t.Fatalf("%q: format=%v key=%q, want format=%v key=%q", tc.env, c.Format, c.APIKey, tc.want, tc.key)
		}
	}
}

func TestFromEnvNemotronDefaultsAndOverrides(t *testing.T) {
	t.Setenv("ARTEX_LLM_PROVIDER", "nemotron")
	t.Setenv("ARTEX_LLM_MODEL", "")
	t.Setenv("ARTEX_LLM_BASE_URL", "")
	t.Setenv("NVIDIA_API_KEY", "nvapi-test")
	t.Setenv("OPENAI_API_KEY", "sk-oai-must-not-be-used")

	c, ok := FromEnv()
	if !ok {
		t.Fatal("FromEnv returned ok=false")
	}
	if c.Format != llm.FormatOpenAI || c.APIKey != "nvapi-test" ||
		c.BaseURL != DefaultNemotronBaseURL || c.Model != defaultNemotronModel {
		t.Fatalf("unexpected Nemotron defaults: %+v", c)
	}

	t.Setenv("ARTEX_LLM_BASE_URL", "https://nim.example/v1")
	t.Setenv("ARTEX_LLM_MODEL", "nvidia/custom-nemotron")
	c, ok = FromEnv()
	if !ok {
		t.Fatal("FromEnv returned ok=false with explicit endpoint and model")
	}
	if c.BaseURL != "https://nim.example/v1" || c.Model != "nvidia/custom-nemotron" {
		t.Fatalf("explicit Nemotron settings were not preserved: base=%q model=%q", c.BaseURL, c.Model)
	}

	t.Setenv("NVIDIA_API_KEY", "")
	if _, ok := FromEnv(); ok {
		t.Fatal("Nemotron configured without NVIDIA_API_KEY should not be usable")
	}
}

func TestFromEnvInfersNemotronFromNVIDIAKey(t *testing.T) {
	t.Setenv("ARTEX_LLM_PROVIDER", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("NVIDIA_API_KEY", "nvapi-test")

	c, ok := FromEnv()
	if !ok || c.Format != llm.FormatOpenAI || c.APIKey != "nvapi-test" {
		t.Fatalf("FromEnv() = (%+v, %v), want Nemotron config from NVIDIA_API_KEY", c, ok)
	}
}
