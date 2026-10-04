// Package unlock 是内容解锁：按等级门槛与前置内容判断哪些内容可以解锁，前置关系不许成环。
//
// 登记拒绝重复 id、未知前置与自环；解锁按玩家增量推进，
// 用就绪堆代替全表扫描，重复查询的代价不随内容数放大。
package unlock

import (
	"container/heap"
	"sort"
)

type node struct {
	level    int
	requires []string
}

// readyItem 是前置已全部解锁、等待等级达标的候选内容。
type readyItem struct {
	id    string
	level int
}

// readyHeap 是按等级门槛排序的最小堆，门槛相同按 id 排序。
type readyHeap []readyItem

func (h readyHeap) Len() int { return len(h) }

func (h readyHeap) Less(i, j int) bool {
	if h[i].level != h[j].level {
		return h[i].level < h[j].level
	}
	return h[i].id < h[j].id
}

func (h readyHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *readyHeap) Push(x any) { *h = append(*h, x.(readyItem)) }

func (h *readyHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

// playerState 是单个玩家的解锁进度。
type playerState struct {
	unlocked map[string]bool
	met      map[string]int // 未就绪内容已解锁的前置数
	ready    readyHeap
	seen     int // 已同步到 addOrder 的位置
}

// Graph 是内容解锁图。
type Graph struct {
	nodes      map[string]node
	dependents map[string][]string // 前置 -> 依赖它的内容
	addOrder   []string            // 登记顺序，供各玩家惰性同步
	players    map[string]*playerState
	scanned    int
}

// New 建空图。
func New() *Graph {
	return &Graph{
		nodes:      map[string]node{},
		dependents: map[string][]string{},
		players:    map[string]*playerState{},
	}
}

// Add 登记一条内容及其等级门槛与前置内容，返回是否登记成功。
// 重复 id、未知前置、自环都拒绝且不改变图；未知前置一律被拒，
// 新节点只有出边没有入边，所以图中不可能出现自环以外的环。
func (g *Graph) Add(id string, level int, requires ...string) bool {
	if _, exists := g.nodes[id]; exists {
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
	g.nodes[id] = node{level: level, requires: append([]string{}, requires...)}
	for _, req := range requires {
		g.dependents[req] = append(g.dependents[req], id)
	}
	g.addOrder = append(g.addOrder, id)
	return true
}

// Unlock 把该玩家新解锁的内容加入台账并返回这批新增的 id（升序）。
// 反复推进到没有新解锁为止；只处理就绪堆里的候选，不全扫内容表。
func (g *Graph) Unlock(player string, level int) []string {
	state := g.players[player]
	if state == nil {
		state = &playerState{unlocked: map[string]bool{}, met: map[string]int{}}
		g.players[player] = state
	}
	state.sync(g)
	out := []string{}
	for len(state.ready) > 0 && state.ready[0].level <= level {
		item := heap.Pop(&state.ready).(readyItem)
		g.scanned++
		if state.unlocked[item.id] {
			continue
		}
		state.unlocked[item.id] = true
		out = append(out, item.id)
		for _, dep := range g.dependents[item.id] {
			if state.unlocked[dep] {
				continue
			}
			state.met[dep]++
			if state.met[dep] == len(g.nodes[dep].requires) {
				delete(state.met, dep)
				heap.Push(&state.ready, readyItem{id: dep, level: g.nodes[dep].level})
			}
		}
	}
	sort.Strings(out)
	return out
}

// sync 把上次查询之后新登记的内容补进该玩家的进度。
func (s *playerState) sync(g *Graph) {
	for s.seen < len(g.addOrder) {
		id := g.addOrder[s.seen]
		s.seen++
		item := g.nodes[id]
		met := 0
		for _, req := range item.requires {
			if s.unlocked[req] {
				met++
			}
		}
		if met == len(item.requires) {
			heap.Push(&s.ready, readyItem{id: id, level: item.level})
		} else if met > 0 {
			s.met[id] = met
		}
	}
}

// Unlocked 返回该玩家已经解锁的内容 id（升序）。
func (g *Graph) Unlocked(player string) []string {
	out := []string{}
	if state := g.players[player]; state != nil {
		for id := range state.unlocked {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

// Scanned 是累计扫描的内容数（规模观测）。
func (g *Graph) Scanned() int { return g.scanned }
