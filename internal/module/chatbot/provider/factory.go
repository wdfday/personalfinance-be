package provider

import (
	"fmt"
	"sync"

	"personalfinancedss/internal/module/chatbot/domain"
)

// providerFactory implements domain.LLMProviderFactory
type providerFactory struct {
	mu              sync.RWMutex
	providers       map[string]domain.LLMProvider
	defaultProvider string
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() domain.LLMProviderFactory {
	f := &providerFactory{
		providers:       make(map[string]domain.LLMProvider),
		defaultProvider: string(domain.ProviderMock),
	}

	// Register mock provider by default
	f.providers[string(domain.ProviderMock)] = NewMockProvider()

	return f
}

func (f *providerFactory) RegisterProvider(providerType string, provider domain.LLMProvider) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.providers[providerType] = provider
}

func (f *providerFactory) SetDefaultProvider(providerType string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.providers[providerType]; !ok {
		return fmt.Errorf("provider %s not registered", providerType)
	}
	f.defaultProvider = providerType
	return nil
}

func (f *providerFactory) GetProvider(providerType string) (domain.LLMProvider, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	provider, ok := f.providers[providerType]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", providerType)
	}
	return provider, nil
}

func (f *providerFactory) GetDefaultProvider() domain.LLMProvider {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.providers[f.defaultProvider]
}

func (f *providerFactory) GetProviderByName(name string) (domain.LLMProvider, error) {
	if name == "" {
		return f.GetDefaultProvider(), nil
	}
	return f.GetProvider(name)
}

func (f *providerFactory) ListProviders() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	names := make([]string, 0, len(f.providers))
	for name := range f.providers {
		names = append(names, name)
	}
	return names
}
