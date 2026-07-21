package cmd

import "testing"

func TestStringListFlagSet(t *testing.T) {
	var values stringListFlag

	if err := values.Set("tmp, dist"); err != nil {
		t.Fatalf("Set gagal: %v", err)
	}
	if err := values.Set("hash.json"); err != nil {
		t.Fatalf("Set gagal: %v", err)
	}

	expected := []string{"tmp", "dist", "hash.json"}
	if len(values) != len(expected) {
		t.Fatalf("jumlah value tidak sesuai: dapat %d, ingin %d", len(values), len(expected))
	}

	for i, value := range values {
		if value != expected[i] {
			t.Fatalf("value ke-%d tidak sesuai: dapat %q, ingin %q", i, value, expected[i])
		}
	}
}
