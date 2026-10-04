package unlock

import "testing"

func TestEmptyGraphUnlocksNothing(t *testing.T) {
	if got := New().Unlock("p", 99); len(got) != 0 {
		t.Fatalf("空图不该解锁任何内容，得到 %v", got)
	}
}

func TestSimpleAddThenUnlock(t *testing.T) {
	graph := New()
	if !graph.Add("a", 1) {
		t.Fatal("登记一条普通内容应该成功")
	}
	if got := graph.Unlock("p", 1); len(got) != 1 || got[0] != "a" {
		t.Fatalf("等级满足时应该解锁 a，得到 %v", got)
	}
}

func TestPrerequisiteBlocksUnlock(t *testing.T) {
	graph := New()
	graph.Add("a", 1)
	graph.Add("b", 1, "a")
	if got := graph.Unlock("p", 1); len(got) != 2 {
		t.Fatalf("前置在链上时应该一起解锁，得到 %v", got)
	}
}

func TestScannedStartsAtZero(t *testing.T) {
	if New().Scanned() != 0 {
		t.Fatal("Scanned 初始应该是 0")
	}
}
