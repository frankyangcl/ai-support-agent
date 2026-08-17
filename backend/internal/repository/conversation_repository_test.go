package repository

import "testing"

func TestTitleFromQuestion(t *testing.T) {
	if got := titleFromQuestion("  Refund policy?  "); got != "Refund policy?" {
		t.Fatalf("got %q", got)
	}
	long := make([]rune, 100)
	for i := range long {
		long[i] = '界'
	}
	if got := titleFromQuestion(string(long)); len([]rune(got)) != 80 {
		t.Fatalf("length %d", len([]rune(got)))
	}
}
func TestEmptyQuestionTitle(t *testing.T) {
	if got := titleFromQuestion("   "); got != "New conversation" {
		t.Fatalf("got %q", got)
	}
}
