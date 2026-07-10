package managementasset

import (
	"bytes"
	"testing"
)

func TestInjectXAIFreeUsageWidget(t *testing.T) {
	original := []byte("<html><body><main>panel</main></body></html>")
	injected := InjectXAIFreeUsageWidget(original)
	if !HasXAIFreeUsageWidget(injected) {
		t.Fatal("page extension marker missing")
	}
	for _, expected := range [][]byte{
		[]byte("cliproxy-xai-free-usage-page"),
		[]byte("#/auth-files?xai-free-usage=1"),
		[]byte("stopImmediatePropagation"),
		[]byte("/v0/management/xai-free-usage"),
		[]byte("enc::v1::"),
		[]byte("MutationObserver"),
		[]byte("visibilitychange"),
		[]byte("30000"),
		[]byte("remaining/limit"),
		[]byte("--cxu-columns"),
		[]byte("cliproxy-xai-free-columns"),
		[]byte("cliproxy-xai-free-rows"),
		[]byte("effectiveColumns"),
		[]byte("rowsPerPage"),
		[]byte("上一页"),
		[]byte("下一页"),
		[]byte("[1,2,3,4,5]"),
		[]byte("最近上游剩余"),
		[]byte("进度条使用最近一次上游额度响应"),
		[]byte("value.remaining"),
		[]byte("上游报告"),
		[]byte("observed_requests"),
		[]byte("最近刷新"),
		[]byte("findSidebar"),
		[]byte("sidebarRect.right"),
		[]byte("history.pushState"),
		[]byte("history.replaceState"),
		[]byte("cliproxy-xai-free-auto-refresh"),
		[]byte("自动刷新"),
		[]byte("autoRefreshEnabled=false"),
	} {
		if !bytes.Contains(injected, expected) {
			t.Fatalf("injected extension missing %q", expected)
		}
	}
	if bytes.Contains(injected, []byte("position:fixed;right:18px;bottom:18px")) {
		t.Fatal("floating widget styles are still present")
	}
	if bytes.Contains(injected, []byte("/v1/chat/completions")) || bytes.Contains(injected, []byte("/v1/responses")) {
		t.Fatal("quota page must not send inference probes")
	}
	if bytes.Contains(injected, []byte("limit-Number(observed")) {
		t.Fatal("quota page must not subtract lifetime observations from the current quota window")
	}
	second := InjectXAIFreeUsageWidget(injected)
	if !bytes.Equal(injected, second) {
		t.Fatal("injection is not idempotent")
	}
}
