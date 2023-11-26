package variant

// Null builtin null
type Null struct{}

// Read implements io.Reader
func (Null) Read(p []byte) (int, error) {
	return 0, nil
}

// Write implements io.Writer
func (Null) Write(p []byte) (int, error) {
	return 0, nil
}

// Size 大小
func (Null) Size() int {
	return 0
}

// Type 类型
func (Null) Type() TypeId {
	return TypeId_Null
}
