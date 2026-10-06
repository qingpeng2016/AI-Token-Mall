package paper

import "testing"

func TestNormalizeSourceCode(t *testing.T) {
	if normalizeSourceCode("semantic-scholar") != sourceSemanticScholar {
		t.Fatal("semantic-scholar alias")
	}
	if normalizeSourceCode(" arxiv ") != sourceArxiv {
		t.Fatal("arxiv trim")
	}
}
