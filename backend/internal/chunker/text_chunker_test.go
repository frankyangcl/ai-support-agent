package chunker

import (
	"strings"
	"testing"
)

func TestSplit(t *testing.T) {

	chunker := NewTextChunker()

	text := strings.Repeat("A", 2500)

	chunks := chunker.Split(text)

	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}

	if chunks[0].CharacterCount != 1000 {
		t.Fatal("chunk0 size incorrect")
	}

	if chunks[1].CharacterCount != 1000 {
		t.Fatal("chunk1 size incorrect")
	}

	if chunks[2].CharacterCount != 800 {
		t.Fatal("chunk2 size incorrect")
	}
}

func TestChineseSplit(t *testing.T) {

	chunker := NewTextChunker()

	text := strings.Repeat("你好世界", 800)

	chunks := chunker.Split(text)

	if len(chunks) == 0 {
		t.Fatal("no chunks")
	}

	for _, c := range chunks {

		if len([]rune(c.Content)) != c.CharacterCount {
			t.Fatal("character count mismatch")
		}
	}
}
