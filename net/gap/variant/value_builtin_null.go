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

// TypeId 类型
func (Null) TypeId() TypeId {
	return TypeId_Null
}

// Indirect 原始值
func (Null) Indirect() any {
	return nil
}
