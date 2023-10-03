package generator
type Generator interface{
	Generate() string
	GenerateChunk() string
}