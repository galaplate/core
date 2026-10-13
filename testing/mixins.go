package testing

type WithMockConfig struct {
	TestCase
}

func (m *WithMockConfig) MockConfig(key string, value any) {
	m.TestCase.MockConfig(key, value)
}

func (m *WithMockConfig) MockConfigMap(values map[string]any) {
	m.TestCase.MockConfigMap(values)
}
