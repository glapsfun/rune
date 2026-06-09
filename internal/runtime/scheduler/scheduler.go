// Package scheduler builds and runs the task dependency DAG. It memoizes each
// (namespace, task, canonical-args) so a node runs at most once per invocation
// (FR-005), runs dependencies before the body and post-hooks after, detects
// cycles, and fails fast on the first error. [parallel] dependencies run
// concurrently (bounded by CPU count) while preserving run-once semantics.
package scheduler

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/rune-task-runner/rune/internal/ast"
)

// Engine resolves dependencies and executes task bodies. The CLI layer
// implements it (it owns the evaluator, parameter binding, and executors).
type Engine interface {
	// ResolveDep evaluates a dependency/post-hook call in the scope of the
	// calling task, returning the target task and its bound parameters.
	ResolveDep(curTask *ast.Task, curParams map[string]string, dep *ast.DepCall) (*ast.Task, map[string]string, error)
	// Execute runs a single task body with its bound parameters.
	Execute(task *ast.Task, params map[string]string) error
	// Namespace returns the memoization namespace for a task (mod path, or "").
	Namespace(task *ast.Task) string
}

// Invocation is a task plus its resolved parameters (a scheduler root).
type Invocation struct {
	Task   *ast.Task
	Params map[string]string
}

// Run executes the given root invocations in order, sharing one memo table so
// repeated tasks run once across the whole invocation.
func Run(engine Engine, roots []Invocation) error {
	s := &state{
		engine:   engine,
		done:     map[string]error{},
		inflight: map[string]*sync.WaitGroup{},
	}
	for _, r := range roots {
		if err := s.run(r.Task, r.Params, nil); err != nil {
			return err
		}
	}
	return nil
}

type state struct {
	engine Engine

	mu       sync.Mutex
	done     map[string]error           // completed keys -> result
	inflight map[string]*sync.WaitGroup // keys currently running
}

// run executes a task once (singleflight by memo key). chain is the
// goroutine-local dependency path of task NAMES, used for cycle detection
// (concurrency-safe because it is passed by value, never shared).
func (s *state) run(task *ast.Task, params map[string]string, chain []string) error {
	for _, name := range chain {
		if name == task.Name {
			return &CycleError{Path: cyclePath(chain, task.Name)}
		}
	}
	key := memoKey(s.engine.Namespace(task), task.Name, params)

	s.mu.Lock()
	if err, ok := s.done[key]; ok {
		s.mu.Unlock()
		return err
	}
	if wg, ok := s.inflight[key]; ok {
		s.mu.Unlock()
		wg.Wait()
		s.mu.Lock()
		err := s.done[key]
		s.mu.Unlock()
		return err
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	s.inflight[key] = wg
	s.mu.Unlock()

	err := s.execute(task, params, append(append([]string{}, chain...), task.Name))

	s.mu.Lock()
	s.done[key] = err
	delete(s.inflight, key)
	s.mu.Unlock()
	wg.Done()
	return err
}

// execute runs dependencies (parallel if [parallel]), the body, then post-hooks.
func (s *state) execute(task *ast.Task, params map[string]string, chain []string) error {
	if task.Attr(ast.AttrParallel) != nil {
		if err := s.runDepsParallel(task, params, task.Deps, chain); err != nil {
			return err
		}
	} else {
		for _, dep := range task.Deps {
			if err := s.runDep(task, params, dep, chain); err != nil {
				return err
			}
		}
	}

	if err := s.engine.Execute(task, params); err != nil {
		return err
	}

	for _, hook := range task.PostHooks {
		if err := s.runDep(task, params, hook, chain); err != nil {
			return err
		}
	}
	return nil
}

func (s *state) runDep(curTask *ast.Task, curParams map[string]string, dep *ast.DepCall, chain []string) error {
	target, depParams, err := s.engine.ResolveDep(curTask, curParams, dep)
	if err != nil {
		return err
	}
	return s.run(target, depParams, chain)
}

// memoKey builds a canonical, order-stable key for a task invocation.
func memoKey(namespace, name string, params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString(namespace)
	b.WriteString("::")
	b.WriteString(name)
	b.WriteByte('(')
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(';')
		}
		fmt.Fprintf(&b, "%s=%s", k, params[k])
	}
	b.WriteByte(')')
	return b.String()
}

// CycleError reports a dependency cycle with the offending path.
type CycleError struct {
	Path []string
}

func (e *CycleError) Error() string {
	return "dependency cycle: " + strings.Join(e.Path, " → ")
}

// cyclePath returns the cycle starting from where `back` first appears in chain.
func cyclePath(chain []string, back string) []string {
	start := 0
	for i, n := range chain {
		if n == back {
			start = i
			break
		}
	}
	return append(append([]string{}, chain[start:]...), back)
}
