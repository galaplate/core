package factory

import (
	"fmt"

	filestorage "github.com/galaplate/core/file-storage"
)

// Factory manages file storage provider registration and access
type Factory struct {
	defaultProvider string
	providers       map[string]filestorage.FileStorageProvider
}

var globalFactory *Factory

// Initialize initializes the global file storage factory with default provider
func Initialize(defaultProvider string, providers map[string]filestorage.FileStorageProvider) {
	globalFactory = &Factory{
		defaultProvider: defaultProvider,
		providers:       providers,
	}
}

// New creates a new Factory instance
func New(defaultProvider string) *Factory {
	return &Factory{
		defaultProvider: defaultProvider,
		providers:       make(map[string]filestorage.FileStorageProvider),
	}
}

// RegisterProvider registers a new storage provider
func (f *Factory) RegisterProvider(name string, provider filestorage.FileStorageProvider) {
	f.providers[name] = provider
}

// GetProvider returns a storage provider by name
func (f *Factory) GetProvider(name string) (filestorage.FileStorageProvider, error) {
	if provider, exists := f.providers[name]; exists {
		return provider, nil
	}
	return nil, fmt.Errorf("provider not found: %s", name)
}

// GetDefaultProvider returns the default storage provider
func (f *Factory) GetDefaultProvider() (filestorage.FileStorageProvider, error) {
	return f.GetProvider(f.defaultProvider)
}

// SetDefaultProvider sets the default storage provider
func (f *Factory) SetDefaultProvider(name string) error {
	if _, err := f.GetProvider(name); err != nil {
		return err
	}
	f.defaultProvider = name
	return nil
}

// Global returns the global factory instance
func Global() *Factory {
	if globalFactory == nil {
		panic("file storage factory not initialized. Call factory.Initialize() during app startup")
	}
	return globalFactory
}

// GetProvider returns the default provider from global factory
func GetProvider() (filestorage.FileStorageProvider, error) {
	return Global().GetDefaultProvider()
}

// SetProvider sets the default provider on global factory
func SetProvider(name string) error {
	return Global().SetDefaultProvider(name)
}

// RegisterProvider registers a provider on the global factory
func RegisterProvider(name string, provider filestorage.FileStorageProvider) {
	Global().RegisterProvider(name, provider)
}
