package config

import "testing"

func TestSaveLoadRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	want := Config{
		SelectionEnabled:    false,
		ShowSelectionSource: false,
		WindowOpacity:       0.42,
		FontOpacity:         0.7,
		Provider:            "google",
		GoogleAPIKey:        "AIza-test-key",
		GoogleAPIBaseURL:    "https://translate.example.com/language/translate/v2",
		APIBaseURL:          "https://mirror.example.com",
	}
	if err := want.Save(); err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.SelectionEnabled != want.SelectionEnabled || got.ShowSelectionSource != want.ShowSelectionSource ||
		got.WindowOpacity != want.WindowOpacity || got.FontOpacity != want.FontOpacity ||
		got.Provider != want.Provider || got.GoogleAPIKey != want.GoogleAPIKey ||
		got.GoogleAPIBaseURL != want.GoogleAPIBaseURL || got.APIBaseURL != want.APIBaseURL {
		t.Fatalf("Load = %+v, want %+v", got, want)
	}
}

func TestLoadDefaultsWhenMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	got, err := Load()
	if err == nil {
		t.Fatal("期望首次加载返回错误（文件不存在）")
	}
	if got.SelectionEnabled != true || got.ShowSelectionSource != true ||
		got.WindowOpacity != 1.0 || got.FontOpacity != 1.0 ||
		got.Provider != "mymemory" || got.GoogleAPIKey != "" {
		t.Fatalf("默认配置错误: %+v", got)
	}
}

func TestClampOpacity(t *testing.T) {
	if ClampOpacity(0.0) != MinOpacity {
		t.Fatalf("ClampOpacity(0.0) = %v, want %v", ClampOpacity(0.0), MinOpacity)
	}
	if ClampOpacity(5.0) != MaxOpacity {
		t.Fatalf("ClampOpacity(5.0) = %v, want %v", ClampOpacity(5.0), MaxOpacity)
	}
	if ClampOpacity(0.55) != 0.55 {
		t.Fatalf("ClampOpacity(0.55) 不应被修改")
	}
}
