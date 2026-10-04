package unlock

import (
	"fmt"
	"testing"
)

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
		t.Fatal("前置里有没登记过的 id 时不该登记成功")
	}
	if !graph.Add("a", 1) {
		t.Fatal("被拒绝的登记不该改变图，a 应该还能重新登记")
	}
	if got := graph.Unlock("p", 1); len(got) != 1 || got[0] != "a" {
		t.Fatalf("重新登记后应该能解锁 a，得到 %v", got)
	}
}

func TestSelfCycleRejected(t *testing.T) {
	graph := New()
	if graph.Add("x", 1, "x") {
		t.Fatal("自环不该登记成功")
	}
	if !graph.Add("x", 1) {
		t.Fatal("被拒绝的自环不该改变图，x 应该还能重新登记")
	}
}

func TestDuplicateAddRejected(t *testing.T) {
	graph := New()
	graph.Add("a", 1)
	graph.Add("q", 1, "a")
	if graph.Add("q", 1) {
		t.Fatal("重复 id 不该登记成功")
	}
	if got := graph.Unlock("p", 1); len(got) != 2 {
		t.Fatalf("被拒绝的重复登记不该丢掉原有前置，得到 %v", got)
	}
}

func TestLevelGateBlocksUnlock(t *testing.T) {
	graph := New()
	graph.Add("raid", 10)
	if got := graph.Unlock("p", 5); len(got) != 0 {
		t.Fatalf("等级低于门槛不该解锁，得到 %v", got)
	}
	if got := graph.Unlock("p", 10); len(got) != 1 || got[0] != "raid" {
		t.Fatalf("等级达到门槛后应该解锁 raid，得到 %v", got)
	}
}

func TestIncrementalUnlockReturnsOnlyNew(t *testing.T) {
	graph := New()
	graph.Add("a", 1)
	graph.Add("b", 1, "a")
	graph.Add("c", 2, "b")
	if got := graph.Unlock("p", 1); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("第一次应该按链推进解锁 a 和 b，得到 %v", got)
	}
	if got := graph.Unlock("p", 1); len(got) != 0 {
		t.Fatalf("已解锁过的内容不该再返回，得到 %v", got)
	}
	if got := graph.Unlock("p", 2); len(got) != 1 || got[0] != "c" {
		t.Fatalf("等级提升后应该只返回新增的 c，得到 %v", got)
	}
	if got := graph.Unlocked("p"); len(got) != 3 {
		t.Fatalf("已解锁集合应该单调不减，得到 %v", got)
	}
}

func TestAddAfterUnlockStillUnlocks(t *testing.T) {
	graph := New()
	graph.Add("a", 1)
	graph.Unlock("p", 1)
	graph.Add("b", 1, "a")
	if got := graph.Unlock("p", 1); len(got) != 1 || got[0] != "b" {
		t.Fatalf("前置已解锁时后登记的内容应该能解锁，得到 %v", got)
	}
}

func TestScannedDoesNotRescan(t *testing.T) {
	graph := New()
	for index := 0; index < 100; index++ {
		graph.Add(fmt.Sprintf("c-%d", index), 1)
	}
	graph.Unlock("p", 1)
	before := graph.Scanned()
	if before > 100 {
		t.Fatalf("首次解锁不该扫描超过内容数，扫描了 %d", before)
	}
	for index := 0; index < 100; index++ {
		graph.Unlock("p", 1)
	}
	if graph.Scanned() != before {
		t.Fatalf("重复查询不该再扫描内容表，扫描数从 %d 涨到 %d", before, graph.Scanned())
	}
}
