package translator

import "testing"

func TestSetProviderSwitchesEngines(t *testing.T) {
	// 顺序存入不同类型的具体实现，验证 atomic.Value 不会因类型不一致而 panic。
	SetProvider(New("mymemory", ""))
	if _, ok := Current().(*mymemory); !ok {
		t.Fatalf("期望 mymemory 引擎, got %T", Current())
	}

	SetProvider(New("google", "some-key"))
	if _, ok := Current().(*google); !ok {
		t.Fatalf("期望 google 引擎, got %T", Current())
	}
}
