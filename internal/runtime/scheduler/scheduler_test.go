package scheduler

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rune-task-runner/rune/internal/ast"
)

// mockEngine is a test Engine: tasks are looked up by name from a map; deps
// carry no real args; Execute records the run order and can fail a named task.
// It is concurrency-safe so the parallel tests can use it.
type mockEngine struct {
	tasks   map[string]*ast.Task
	failOn  string
	execErr error
	delay   time.Duration

	mu      sync.Mutex
	order   []string
	running int
	maxConc int
}

func (m *mockEngine) ResolveDep(_ *ast.Task, _ map[string]string, dep *ast.DepCall) (*ast.Task, map[string]string, error) {
	t, ok := m.tasks[dep.Name]
	if !ok {
		return nil, nil, errors.New("unknown task: " + dep.Name)
	}
	return t, map[string]string{}, nil
}

func (m *mockEngine) Execute(task *ast.Task, _ map[string]string) error {
	m.mu.Lock()
	m.order = append(m.order, task.Name)
	m.running++
	if m.running > m.maxConc {
		m.maxConc = m.running
	}
	m.mu.Unlock()

	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	m.mu.Lock()
	m.running--
	m.mu.Unlock()

	if task.Name == m.failOn {
		return m.execErr
	}
	return nil
}

func (m *mockEngine) Namespace(_ *ast.Task) string { return "" }

func task(name string, deps ...string) *ast.Task {
	t := &ast.Task{Name: name}
	for _, d := range deps {
		t.Deps = append(t.Deps, &ast.DepCall{Name: d})
	}
	return t
}

func newEngine(tasks ...*ast.Task) *mockEngine {
	m := &mockEngine{tasks: map[string]*ast.Task{}}
	for _, t := range tasks {
		m.tasks[t.Name] = t
	}
	return m
}

func TestSchedulerTopoOrder(t *testing.T) {
	// build depends on greet; greet runs first.
	greet := task("greet")
	build := task("build", "greet")
	m := newEngine(greet, build)
	if err := Run(m, []Invocation{{Task: build, Params: map[string]string{}}}); err != nil {
		t.Fatal(err)
	}
	want := []string{"greet", "build"}
	if strings.Join(m.order, ",") != strings.Join(want, ",") {
		t.Errorf("order = %v, want %v", m.order, want)
	}
}

func TestSchedulerRunOnce(t *testing.T) {
	// Diamond: top -> {a, b} -> base. base must run exactly once.
	base := task("base")
	a := task("a", "base")
	b := task("b", "base")
	top := task("top", "a", "b")
	m := newEngine(base, a, b, top)
	if err := Run(m, []Invocation{{Task: top, Params: map[string]string{}}}); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, n := range m.order {
		if n == "base" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("base ran %d times, want 1 (order=%v)", count, m.order)
	}
}

func TestSchedulerFailFast(t *testing.T) {
	// a -> b(fails) -> c ; c must never run, a must never run.
	c := task("c")
	b := task("b", "c")
	a := task("a", "b")
	m := newEngine(a, b, c)
	m.failOn = "b"
	m.execErr = errors.New("boom")
	err := Run(m, []Invocation{{Task: a, Params: map[string]string{}}})
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err = %v, want boom", err)
	}
	for _, n := range m.order {
		if n == "a" {
			t.Error("a ran but its dependency b failed")
		}
	}
}

func TestSchedulerCycleDetection(t *testing.T) {
	// a -> b -> a (runtime guard; analyzer normally catches this first).
	a := task("a", "b")
	b := task("b", "a")
	m := newEngine(a, b)
	err := Run(m, []Invocation{{Task: a, Params: map[string]string{}}})
	var ce *CycleError
	if !errors.As(err, &ce) {
		t.Fatalf("err = %v, want CycleError", err)
	}
}

func TestSchedulerMemoAcrossRoots(t *testing.T) {
	// Two roots that share a dependency: it runs once total.
	base := task("base")
	a := task("a", "base")
	b := task("b", "base")
	m := newEngine(base, a, b)
	roots := []Invocation{{Task: a, Params: map[string]string{}}, {Task: b, Params: map[string]string{}}}
	if err := Run(m, roots); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, n := range m.order {
		if n == "base" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("base ran %d times across roots, want 1", count)
	}
}
