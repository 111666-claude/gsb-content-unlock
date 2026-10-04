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

func TestUnknownPrerequisiteRejected(t *testing.T) {
	graph := New()
	if graph.Add("a", 1, "ghost") {
		t.Fatal("前置里有没登记过的 id 应该登记失败")
	}
	if got := graph.Unlock("p", 99); len(got) != 0 {
		t.Fatalf("登记失败的条目不改变图，得到 %v", got)
	}
}

func TestSelfCycleRejected(t *testing.T) {
	graph := New()
	if graph.Add("x", 1, "x") {
		t.Fatal("自环应该登记失败")
	}
	graph.Add("a", 1)
	graph.Add("b", 1, "a")
	if graph.Add("a", 1, "b") {
		t.Fatal("重复 id 应该登记失败")
	}
	if !graph.Add("c", 1, "b") {
		t.Fatal("正常链上追加应该成功")
	}
	if graph.Add("d", 1, "c") != true {
		t.Fatal("无环链应该登记成功")
	}
}

func TestIndirectCycleRejected(t *testing.T) {
	graph := New()
	graph.Add("a", 1)
	graph.Add("b", 1, "a")
	if graph.Add("a2", 1, "b", "a2") {
		t.Fatal("自环应该登记失败")
	}
	if got := graph.Unlock("p", 1); len(got) != 2 {
		t.Fatalf("被拒绝的登记不改变图，得到 %v", got)
	}
}

func TestLevelGateBlocksUnlock(t *testing.T) {
	graph := New()
	graph.Add("raid", 10)
	if got := graph.Unlock("p", 5); len(got) != 0 {
		t.Fatalf("等级不够不该解锁，得到 %v", got)
	}
	if got := graph.Unlock("p", 10); len(got) != 1 || got[0] != "raid" {
		t.Fatalf("等级达到门槛应该解锁，得到 %v", got)
	}
}

func TestIncrementalUnlockReturnsOnlyNew(t *testing.T) {
	graph := New()
	graph.Add("a", 1)
	graph.Add("b", 2, "a")
	if got := graph.Unlock("p", 1); len(got) != 1 || got[0] != "a" {
		t.Fatalf("第一次只该解锁 a，得到 %v", got)
	}
	if got := graph.Unlock("p", 1); len(got) != 0 {
		t.Fatalf("已解锁过的不该再返回，得到 %v", got)
	}
	if got := graph.Unlock("p", 2); len(got) != 1 || got[0] != "b" {
		t.Fatalf("升级后只该返回新增的 b，得到 %v", got)
	}
	if got := graph.Unlocked("p"); len(got) != 2 {
		t.Fatalf("台账应该累计两条，得到 %v", got)
	}
}

func TestDuplicateAddRejected(t *testing.T) {
	graph := New()
	graph.Add("a", 1)
	graph.Add("q", 1, "a")
	if graph.Add("q", 1) {
		t.Fatal("重复 id 应该登记失败")
	}
	if got := graph.Unlock("p", 1); len(got) != 2 {
		t.Fatalf("重复登记不改变图，q 仍依赖 a，得到 %v", got)
	}
}
