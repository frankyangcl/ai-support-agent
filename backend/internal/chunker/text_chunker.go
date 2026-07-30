package chunker

type TextChunk struct {
	Index          int
	Content        string
	CharacterCount int
}

type TextChunker struct {
	ChunkSize int
	Overlap   int
}

func NewTextChunker() *TextChunker {
	return &TextChunker{
		ChunkSize: 1000,
		Overlap:   150,
	}
}

func (c *TextChunker) Split(text string) []TextChunk {
	runes := []rune(text)

	if len(runes) == 0 {
		return nil
	}

	var chunks []TextChunk

	start := 0
	index := 0

	for start < len(runes) {

		end := start + c.ChunkSize
		if end > len(runes) {
			end = len(runes)
		}

		content := string(runes[start:end])

		chunks = append(chunks, TextChunk{
			Index:          index,
			Content:        content,
			CharacterCount: len([]rune(content)),
		})

		if end == len(runes) {
			break
		}

		start = end - c.Overlap
		index++
	}

	return chunks
}
