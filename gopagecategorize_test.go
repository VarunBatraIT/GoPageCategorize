package gopagecategorize

import (
	"testing"
)

func TestAnalyzeUrl(t *testing.T) {
	scores, err := AnalyzeUrl("https://github.com/varunbatrait")
	if len(scores) == 0 {
		t.Errorf("Failed in test")
	}
	if err != nil {
		t.Error(err)
	}
}
