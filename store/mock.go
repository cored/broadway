package store

// FakeStore mock for the store
type FakeStore struct {
	MockSetValue func(path, value string) error
	MockValue    func(path string) string
	MockValues   func(path string) map[string]string
	MockDelete   func(path string) error
}

// SetValue mocked implementation
func (fs *FakeStore) SetValue(path, value string) error {
	return fs.MockSetValue(path, value)
}

// Value mocked implementation
func (fs *FakeStore) Value(path string) string {
	return fs.MockValue(path)
}

// Values mocked implementation
func (fs *FakeStore) Values(path string) map[string]string {
	return fs.MockValues(path)
}

// Delete mocked implementation
func (fs *FakeStore) Delete(path string) error {
	return fs.MockDelete(path)
}
