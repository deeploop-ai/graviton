package auth

import "testing"

func TestWeChatIdentityUID(t *testing.T) {
	t.Parallel()
	if got := WeChatIdentityUID("union-1", "openid-1"); got != "union-1" {
		t.Fatalf("expected unionid, got %q", got)
	}
	if got := WeChatIdentityUID("", "openid-2"); got != "openid-2" {
		t.Fatalf("expected openid fallback, got %q", got)
	}
}

func TestIsWeChatProvider(t *testing.T) {
	t.Parallel()
	for _, p := range []string{ProviderWeChatWeb, ProviderWeChatMP, ProviderWeChatMiniProgram, ProviderWeChatApp} {
		if !IsWeChatProvider(p) {
			t.Fatalf("expected %q to be wechat provider", p)
		}
	}
	if IsWeChatProvider("google") {
		t.Fatal("google should not be wechat provider")
	}
}
