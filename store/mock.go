package store

type FakeStore struct {
	MockSetValue func(path, value string) error
	MockValue    func(path string) string
	MockValues   func(path string) map[string]string
	MockDelete   func(path string) error
}

func (fs *FakeStore) SetValue(path, value string) error {
	return fs.MockSetValue(path, value)
}

func (fs *FakeStore) Value(path string) string {
	return fs.MockValue(path)
}

func (fs *FakeStore) Values(path string) map[string]string {
	return fs.MockValues(path)
}

func (fs *FakeStore) Delete(path string) error {
	return fs.MockDelete(path)
}
