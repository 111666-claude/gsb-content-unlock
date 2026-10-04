// Package unlock 是内容解锁：按等级门槛与前置内容判断哪些内容可以解锁，前置关系不许成环。
package unlock

import "sort"

type node struct {
	level    int
	requires []string
}

type progress struct {
	unlocked  map[string]bool
	remaining map[string]int
	queue     []string
	level     int
	started   bool
}

// Graph 是内容解锁图。
type Graph struct {
	nodes      map[string]node
	dependents map[string][]string
	byLevel    map[int][]string
	levels     []int
	players    map[string]*progress
	scanned    int
}

// New 建空图。
func New() *Graph {
	return &Graph{
		nodes:      map[string]node{},
		dependents: map[string][]string{},
		byLevel:    map[int][]string{},
		players:    map[string]*progress{},
	}
}

// Add 登记一条内容及其等级门槛与前置内容，返回是否登记成功。
// 重复 id、未知前置、会成环（含自环）都返回 false 且不改变图。
func (g *Graph) Add(id string, level int, requires ...string) bool {
	if _, dup := g.nodes[id]; dup {
		return false
	}
	for _, req := range requires {
		if req == id {
			return false
		}
		if _, ok := g.nodes[req]; !ok {
			return false
		}
	}
	if g.createsCycle(id, requires) {
		return false
	}

	g.nodes[id] = node{level: level, requires: append([]string{}, requires...)}
	for _, req := range requires {
		g.dependents[req] = append(g.dependents[req], id)
	}
	if _, ok := g.byLevel[level]; !ok {
		pos := sort.SearchInts(g.levels, level)
		g.levels = append(g.levels, 0)
		copy(g.levels[pos+1:], g.levels[pos:])
		g.levels[pos] = level
	}
	g.byLevel[level] = append(g.byLevel[level], id)
	for _, p := range g.players {
		if p.started && level <= p.level {
			g.account(p, id)
		}
	}
	return true
}

// createsCycle 判断以 requires 为前置加入 id 是否会成环：
// 任一前置能沿 requires 边到达 id 即成环。
func (g *Graph) createsCycle(id string, requires []string) bool {
	stack := append([]string{}, requires...)
	seen := map[string]bool{}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if cur == id {
			return true
		}
		if seen[cur] {
			continue
		}
		seen[cur] = true
		stack = append(stack, g.nodes[cur].requires...)
	}
	return false
}

// account 为玩家统计 id 还差几个前置，够了就进待解锁队列。
func (g *Graph) account(p *progress, id string) {
	g.scanned++
	locked := 0
	for _, req := range g.nodes[id].requires {
		if !p.unlocked[req] {
			locked++
		}
	}
	p.remaining[id] = locked
	if locked == 0 {
		p.queue = append(p.queue, id)
	}
}

// Unlock 反复推进到没有新解锁为止，只返回该玩家本次新增的 id（升序）。
func (g *Graph) Unlock(player string, level int) []string {
	p := g.players[player]
	if p == nil {
		p = &progress{unlocked: map[string]bool{}, remaining: map[string]int{}}
		g.players[player] = p
	}

	low := 0
	if p.started {
		low = sort.SearchInts(g.levels, p.level+1)
	}
	if level > p.level {
		p.level = level
	}
	p.started = true
	high := sort.SearchInts(g.levels, p.level+1)
	for ; low < high; low++ {
		for _, id := range g.byLevel[g.levels[low]] {
			g.account(p, id)
		}
	}
	newly := []string{}
	for len(p.queue) > 0 {
		id := p.queue[0]
		p.queue = p.queue[1:]
		g.scanned++
		if p.unlocked[id] {
			continue
		}
		p.unlocked[id] = true
		newly = append(newly, id)
		for _, dep := range g.dependents[id] {
			rem, ok := p.remaining[dep]
			if !ok {
				continue
			}
			rem--
			p.remaining[dep] = rem
			if rem == 0 {
				p.queue = append(p.queue, dep)
			}
		}
	}
	sort.Strings(newly)
	return newly
}

// Unlocked 返回该玩家已经解锁的内容 id（升序）。
func (g *Graph) Unlocked(player string) []string {
	out := []string{}
	for id := range g.players[player].unlocked {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// Scanned 是累计扫描的内容数（规模观测）。
func (g *Graph) Scanned() int { return g.scanned }
