package generator
type Generator interface{
	Start() []byte
	Generate() []byte
	GenerateChunk() []byte
	ChunkSeparator() []byte
	End() []byte
}