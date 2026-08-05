package scheduler

import (
	"reflect"
	"testing"
)

func TestTruncate(t *testing.T) {
	if got := truncate("hello world", 5); got != "hello" {
		t.Errorf("expected 'hello', got '%s'", got)
	}
	if got := truncate("hi", 5); got != "hi" {
		t.Errorf("expected 'hi', got '%s'", got)
	}
}

func TestItoa(t *testing.T) {
	if got := itoa(42); got != "42" {
		t.Errorf("expected '42', got '%s'", got)
	}
	if got := itoa(-100); got != "-100" {
		t.Errorf("expected '-100', got '%s'", got)
	}
}

func TestCleanText(t *testing.T) {
	if got := cleanText("... Hello world"); got != "Hello world" {
		t.Errorf("expected 'Hello world', got '%s'", got)
	}
	if got := cleanText("---   Sample tweet   "); got != "Sample tweet" {
		t.Errorf("expected 'Sample tweet', got '%s'", got)
	}

	longText := stringsRepeat("a", 300)
	cleaned := cleanText(longText)
	if len([]rune(cleaned)) != 280 {
		t.Errorf("expected 280 runes, got %d", len([]rune(cleaned)))
	}
}

func TestDedupeInts(t *testing.T) {
	in := []int64{1, 2, 2, 3, 1, 4, 3, 5}
	expected := []int64{1, 2, 3, 4, 5}
	got := dedupeInts(in)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func stringsRepeat(s string, count int) string {
	b := make([]byte, 0, len(s)*count)
	for i := 0; i < count; i++ {
		b = append(b, s...)
	}
	return string(b)
}
