package factory

import (
	"fmt"
	"mime/multipart"

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

// Upload uploads a file using the default provider
func (f *Factory) Upload(file *multipart.FileHeader) filestorage.UploadMetadata {
	provider, err := f.GetDefaultProvider()
	if err != nil {
		return filestorage.UploadMetadata{
			Error: fmt.Sprintf("failed to get default provider: %v", err),
		}
	}
	return provider.Upload(file)
}

// UploadWith uploads a file using a specific provider
func (f *Factory) UploadWith(providerName string, file *multipart.FileHeader) filestorage.UploadMetadata {
	provider, err := f.GetProvider(providerName)
	if err != nil {
		return filestorage.UploadMetadata{
			Error: fmt.Sprintf("failed to get provider %s: %v", providerName, err),
		}
	}
	return provider.Upload(file)
}

// Delete deletes a file using the default provider
func (f *Factory) Delete(filePath string, storageType string) error {
	provider, err := f.GetDefaultProvider()
	if err != nil {
		return fmt.Errorf("failed to get default provider: %w", err)
	}
	return provider.Delete(filePath, storageType)
}

// DeleteWith deletes a file using a specific provider
func (f *Factory) DeleteWith(providerName string, filePath string, storageType string) error {
	provider, err := f.GetProvider(providerName)
	if err != nil {
		return fmt.Errorf("failed to get provider %s: %w", providerName, err)
	}
	return provider.Delete(filePath, storageType)
}

// GetDownloadURL returns a download URL using the default provider
func (f *Factory) GetDownloadURL(metadata filestorage.UploadMetadata) string {
	provider, err := f.GetDefaultProvider()
	if err != nil {
		return ""
	}
	return provider.GetDownloadURL(metadata)
}

// GetDownloadURLWith returns a download URL using a specific provider
func (f *Factory) GetDownloadURLWith(providerName string, metadata filestorage.UploadMetadata) string {
	provider, err := f.GetProvider(providerName)
	if err != nil {
		return ""
	}
	return provider.GetDownloadURL(metadata)
}

// ValidateFileExists validates if a file exists using the default provider
func (f *Factory) ValidateFileExists(metadata filestorage.UploadMetadata) bool {
	provider, err := f.GetDefaultProvider()
	if err != nil {
		return false
	}
	return provider.ValidateFileExists(metadata)
}

// ValidateFileExistsWith validates if a file exists using a specific provider
func (f *Factory) ValidateFileExistsWith(providerName string, metadata filestorage.UploadMetadata) bool {
	provider, err := f.GetProvider(providerName)
	if err != nil {
		return false
	}
	return provider.ValidateFileExists(metadata)
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

// Upload uploads a file using the default provider from global factory
func Upload(file *multipart.FileHeader) filestorage.UploadMetadata {
	return Global().Upload(file)
}

// UploadWith uploads a file using a specific provider from global factory
func UploadWith(providerName string, file *multipart.FileHeader) filestorage.UploadMetadata {
	return Global().UploadWith(providerName, file)
}

// Delete deletes a file using the default provider from global factory
func Delete(filePath string, storageType string) error {
	return Global().Delete(filePath, storageType)
}

// DeleteWith deletes a file using a specific provider from global factory
func DeleteWith(providerName string, filePath string, storageType string) error {
	return Global().DeleteWith(providerName, filePath, storageType)
}

// GetDownloadURL returns a download URL using the default provider from global factory
func GetDownloadURL(metadata filestorage.UploadMetadata) string {
	return Global().GetDownloadURL(metadata)
}

// GetDownloadURLWith returns a download URL using a specific provider from global factory
func GetDownloadURLWith(providerName string, metadata filestorage.UploadMetadata) string {
	return Global().GetDownloadURLWith(providerName, metadata)
}

// ValidateFileExists validates if a file exists using the default provider from global factory
func ValidateFileExists(metadata filestorage.UploadMetadata) bool {
	return Global().ValidateFileExists(metadata)
}

// ValidateFileExistsWith validates if a file exists using a specific provider from global factory
func ValidateFileExistsWith(providerName string, metadata filestorage.UploadMetadata) bool {
	return Global().ValidateFileExistsWith(providerName, metadata)
}
