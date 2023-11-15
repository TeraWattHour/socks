package socks

import (
	"fmt"
	"testing"
)

func TestBasicEvaluation(t *testing.T) {
	s, err := NewSocks("test_data/basic")
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	res, err := s.Run("nested.html", map[string]interface{}{
		"Phrases": []string{"Herzlich willkommen", "Willkommen"},
	})
	if err != nil {
		t.Errorf("Expected no error, got %sock", err)
	}

	fmt.Println(res)
}
