package crud

import "testing"

func TestEncodeDecodeCursorPageToken(t *testing.T) {
	token := EncodeCursorPageToken(40, 20, "created_at desc", "abc", map[string]string{"created_at": "2026-01-01T00:00:00Z"}, "sum")
	data, err := DecodePageTokenFull(token)
	if err != nil {
		t.Fatalf("DecodePageTokenFull err=%v", err)
	}
	if data.Mode != TokenModeCursor {
		t.Fatalf("mode=%s", data.Mode)
	}
	if data.Offset != 40 || data.PageSize != 20 {
		t.Fatalf("unexpected offset/page_size: %d/%d", data.Offset, data.PageSize)
	}
	if data.OrderBy != "created_at desc" || data.FilterDigest != "abc" || data.Checksum != "sum" {
		t.Fatalf("unexpected cursor metadata: %+v", data)
	}
}

func TestSignedCursorPageToken(t *testing.T) {
	const secret = "test-secret"
	token, err := EncodeSignedCursorPageToken(10, 5, "created_at desc", "fd", nil, "cs", secret)
	if err != nil {
		t.Fatalf("EncodeSignedCursorPageToken err=%v", err)
	}
	if _, err := DecodeSignedPageTokenFull(token, secret); err != nil {
		t.Fatalf("DecodeSignedPageTokenFull err=%v", err)
	}
	if _, err := DecodeSignedPageTokenFull(token, "bad-secret"); err == nil {
		t.Fatalf("expected signature validation failure")
	}
}
