package typedb

// DeserializeForBenchmark exports deserializeForType for benchmark testing.
// This allows benchmarks to call deserialize functions directly without
// going through the public API, isolating deserialization overhead.
func DeserializeForBenchmark[T ModelInterface](row map[string]any) (T, error) {
	return deserializeForType[T](row)
}

// DeserializeForBenchmarkDirect exports deserialize for benchmark testing.
// This allows benchmarks to call deserialize directly on an existing model instance.
func DeserializeForBenchmarkDirect(row map[string]any, dest ModelInterface) error {
	return deserialize(row, dest)
}
