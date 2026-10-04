// Package unlock 是内容解锁：按等级门槛与前置内容判断哪些内容可以解锁，前置关系不许成环。
// 缺陷：自环不拦、未知前置照收、重复登记照收、不看等级门槛、已解锁的也会再返回一次、
// 每次查询全扫内容表。
package unlock

import "sort"

type node struct {
	level    int
	requires []string
}

// Graph 是内容解锁图。
type Graph struct {
	nodes    map[string]node
	order    []string
	unlocked map[string]map[string]bool
	scanned  int
}

// New 建空图。
func New() *Graph {
	return &Graph{nodes: map[string]node{}, unlocked: map[string]map[string]bool{}}
}

// Add 登记一条内容及其等级门槛与前置内容，返回是否登记成功。
func (g *Graph) Add(id string, level int, requires ...string) bool {
	g.nodes[id] = node{level: level, requires: append([]string{}, requires...)}
	g.order = append(g.order, id)
	return true
}

// Unlock 把该玩家新解锁的内容加入台账并返回这批新增的 id（升序）。
func (g *Graph) Unlock(player string, level int) []string {
	g.scanned += len(g.nodes)
	set := g.unlocked[player]
	if set == nil {
		set = map[string]bool{}
		g.unlocked[player] = set
	}
	out := []string{}
	for _, id := range g.order {
		item := g.nodes[id]
		ready := true
		for _, need := range item.requires {
			if !set[need] {
				ready = false
				break
			}
		}
		if ready {
			set[id] = true
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

// Unlocked 返回该玩家已经解锁的内容 id（升序）。
func (g *Graph) Unlocked(player string) []string {
	out := []string{}
	for id := range g.unlocked[player] {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// Scanned 是累计扫描的内容数（规模观测）。
func (g *Graph) Scanned() int { return g.scanned }
