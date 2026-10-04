// Command content-unlock 跑内容解锁样例。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/unlock"
)

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("content-unlock", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sample := flags.String("sample", "unknown", "unknown / cycle / level / again / dup / scan")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	switch *sample {
	case "unknown":
		graph := unlock.New()
		fmt.Fprintf(stdout, "ok=%v\n", graph.Add("a", 1, "ghost"))
	case "cycle":
		graph := unlock.New()
		fmt.Fprintf(stdout, "ok=%v\n", graph.Add("x", 1, "x"))
	case "level":
		graph := unlock.New()
		graph.Add("raid", 10)
		fmt.Fprintf(stdout, "count=%d\n", len(graph.Unlock("p", 5)))
	case "again":
		graph := unlock.New()
		graph.Add("a", 1)
		graph.Unlock("p", 1)
		fmt.Fprintf(stdout, "second=%d\n", len(graph.Unlock("p", 1)))
	case "dup":
		graph := unlock.New()
		graph.Add("a", 1)
		graph.Add("q", 1, "a")
		second := graph.Add("q", 1)
		fmt.Fprintf(stdout, "second=%v\n", second)
	case "scan":
		graph := unlock.New()
		for index := 0; index < 3000; index++ {
			graph.Add(fmt.Sprintf("c-%d", index), 1)
		}
		for index := 0; index < 3000; index++ {
			graph.Unlock("p-0", 1)
		}
		fmt.Fprintf(stdout, "scanned=%d\n", graph.Scanned())
	default:
		fmt.Fprintln(stderr, "需要 --sample unknown|cycle|level|again|dup|scan")
		return 2
	}
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
