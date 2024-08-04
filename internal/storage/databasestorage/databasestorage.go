package databasestorage

type Storage struct {
}

func NewStorage() Storage {
	return Storage{}
}

func (s *Storage) SetValue(metricType, name, value string) error {
	return nil
}

func (s *Storage) GetValue(metricType, name string) (string, error) {
	return "", nil
}

func (m *Storage) GetRows() []string {
	return nil
}
