package timer

import "testing"

func TestDeriveToken(t *testing.T) {
	tok := deriveToken("test123")
	if tok == "" {
		t.Fatal("token 不应为空")
	}
	// 相同密码派生相同 token（确定性）
	if deriveToken("test123") != tok {
		t.Fatal("相同密码应派生相同 token")
	}
	// 不同密码派生不同 token
	if deriveToken("other456") == tok {
		t.Fatal("不同密码应派生不同 token")
	}
	// token 是 hex 编码（偶数长度）
	if len(tok)%2 != 0 {
		t.Fatalf("token 应为 hex 编码, 长度异常: %d", len(tok))
	}
}
