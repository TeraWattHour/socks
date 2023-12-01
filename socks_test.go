package socks

import (
	"fmt"
	"testing"
)

func TestBasicEvaluation(t *testing.T) {
	s, err := NewSocks("test_data", map[string]interface{}{
		"Server": "Frankfurt",
	})
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	s.AddGlobal("now", "2019-01-01")

	res, err := s.Run("nested.html", map[string]interface{}{
		"Phrases": []string{"Herzlich willkommen", "Willkommen"},
	})
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	fmt.Println("result", res)
}
