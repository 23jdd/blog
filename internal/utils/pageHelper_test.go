package utils

import "testing"

func TestResolveOffsetLimit_DefaultAndClamp(t *testing.T) {
	offset, limit := ResolveOffsetLimit(0, 0)
	if offset != 0 {
		t.Fatalf("offset should be 0, got %d", offset)
	}
	if limit != DefaultPageSize {
		t.Fatalf("limit should be %d, got %d", DefaultPageSize, limit)
	}

	offset, limit = ResolveOffsetLimit(999999, 1000)
	if limit != MaxPageSize {
		t.Fatalf("limit should clamp to %d, got %d", MaxPageSize, limit)
	}
	if offset != MaxOffset {
		t.Fatalf("offset should clamp to %d, got %d", MaxOffset, offset)
	}
}
