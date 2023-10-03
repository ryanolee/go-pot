package chaff

type (
	Reference struct {
		Path string
		Generator Generator
		SchemaNode SchemaNode
	}

	ReferenceHandler struct{
		CurrentPath string
		References map[string]Reference
		Errors map[string]error
	}
)

func NewReferenceHandler() ReferenceHandler {
	return ReferenceHandler{
		CurrentPath: "#",
		References: make(map[string]Reference),
		Errors: make(map[string]error),
	}
}

func (h *ReferenceHandler) ParseNodeInScope(scope string , node SchemaNode, metadata ParserMetadata) (Generator, error) {
	h.PushToPath(scope)
	generator, err := ParseNode(node, metadata)
	h.PopFromPath(scope)
	return generator, err
}

func (h *ReferenceHandler) PushToPath(pathPart string) {
	h.CurrentPath += pathPart
}

func (h *ReferenceHandler) PopFromPath(pathPart string) {
	h.CurrentPath = h.CurrentPath[:len(h.CurrentPath)-len(pathPart)]
}

func (h *ReferenceHandler) AddReference(node SchemaNode, generator Generator) {
	h.AddIdReference(h.CurrentPath, node, generator)
}

func (h *ReferenceHandler) AddIdReference(path string, node SchemaNode, generator Generator){
	h.References[path] = Reference{
		Path: path,
		SchemaNode: node,
		Generator: generator,
	}
}

func (h *ReferenceHandler) HandleError(err error) {
	h.Errors[h.CurrentPath] = err
}

func (h *ReferenceHandler) Lookup(path string) (Reference, bool) {
	Reference, ok := h.References[path]
	return Reference, ok
}