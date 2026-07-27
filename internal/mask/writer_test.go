package mask

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

func newTestWriter(env ...string) (*Writer, *bytes.Buffer) {
	var buf bytes.Buffer
	return NewWriter(&buf, NewSet(env, nil, nil)), &buf
}

func TestWriter_BasicReplacement(t *testing.T) {
	w, buf := newTestWriter("API_TOKEN=hunter2-xyz")
	n, err := w.Write([]byte("token is hunter2-xyz\n"))
	if err != nil {
		t.Fatal(err)
	}
	if n != len("token is hunter2-xyz\n") {
		t.Errorf("Write returned n=%d, want full input length", n)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "token is "+Placeholder+"\n" {
		t.Errorf("output = %q", got)
	}
}

func TestWriter_MultipleOccurrencesAcrossLines(t *testing.T) {
	w, buf := newTestWriter("API_TOKEN=hunter2-xyz")
	_, _ = w.Write([]byte("a hunter2-xyz b\nhunter2-xyz\nc hunter2-xyz"))
	_ = w.Flush()
	want := "a " + Placeholder + " b\n" + Placeholder + "\nc " + Placeholder
	if got := buf.String(); got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestWriter_ChunkSplitByteAtATime(t *testing.T) {
	w, buf := newTestWriter("API_TOKEN=hunter2-xyz")
	in := "start hunter2-xyz middle hunter2-xyz end\n"
	for i := 0; i < len(in); i++ {
		if _, err := w.Write([]byte{in[i]}); err != nil {
			t.Fatal(err)
		}
	}
	_ = w.Flush()
	want := "start " + Placeholder + " middle " + Placeholder + " end\n"
	if got := buf.String(); got != want {
		t.Errorf("byte-at-a-time output = %q, want %q", got, want)
	}
}

func TestWriter_ChunkSplitTwoWrites(t *testing.T) {
	w, buf := newTestWriter("API_TOKEN=hunter2-xyz")
	_, _ = w.Write([]byte("prefix hunter2-"))
	// The tail "hunter2-" is a proper prefix of the secret: it must be withheld.
	if got := buf.String(); got != "prefix " {
		t.Errorf("after first write, emitted %q; carry not withheld", got)
	}
	_, _ = w.Write([]byte("xyz suffix"))
	if got := buf.String(); got != "prefix "+Placeholder+" suffix" {
		t.Errorf("after completing write, output = %q", got)
	}
}

func TestWriter_CarryBounded(t *testing.T) {
	// The carry may never exceed maxEntryLen-1 bytes: writing a long run of a
	// non-matching-but-first-byte-alike stream must not buffer unboundedly.
	w, buf := newTestWriter("API_TOKEN=hunter2-xyz")
	in := strings.Repeat("h", 4096)
	_, _ = w.Write([]byte(in))
	if emitted := buf.Len(); emitted < len(in)-(len("hunter2-xyz")-1) {
		t.Errorf("emitted %d of %d bytes; carry exceeds maxEntryLen-1", emitted, len(in))
	}
	_ = w.Flush()
	if got := buf.String(); got != in {
		t.Errorf("non-secret bytes were altered")
	}
}

func TestWriter_FlushEmitsIncompletePrefixVerbatim(t *testing.T) {
	w, buf := newTestWriter("API_TOKEN=hunter2-xyz")
	_, _ = w.Write([]byte("tail hunter2-"))
	_ = w.Flush()
	// An incomplete prefix is not the secret; after flush it is emitted as-is.
	if got := buf.String(); got != "tail hunter2-" {
		t.Errorf("flush output = %q", got)
	}
	// And the carry is cleared: later writes start fresh.
	_, _ = w.Write([]byte("xyz"))
	_ = w.Flush()
	if got := buf.String(); got != "tail hunter2-xyz" {
		t.Errorf("post-flush write mangled: %q", got)
	}
}

func TestWriter_FlushMasksCompletedEntryInCarry(t *testing.T) {
	// When one secret is a prefix of a longer one, the writer withholds the
	// ambiguous tail. If the stream ends there, Flush must still mask the
	// shorter, fully-present secret rather than emit it verbatim.
	w, buf := newTestWriter("A_TOKEN=hunter2-xyz", "B_TOKEN=hunter2-xyz-extended")
	_, _ = w.Write([]byte("end hunter2-xyz"))
	_ = w.Flush()
	if got := buf.String(); got != "end "+Placeholder {
		t.Errorf("flush leaked a completed shorter secret: %q", got)
	}
}

func TestWriter_NestedSecretsNoFragment(t *testing.T) {
	w, buf := newTestWriter("A_TOKEN=hunter2-xyz", "B_TOKEN=hunter2-xyz-extended")
	_, _ = w.Write([]byte("x hunter2-xyz-extended y\n"))
	_ = w.Flush()
	if got := buf.String(); got != "x "+Placeholder+" y\n" {
		t.Errorf("nested secrets leaked a fragment: %q", got)
	}
}

func TestWriter_CloseFlushes(t *testing.T) {
	w, buf := newTestWriter("API_TOKEN=hunter2-xyz")
	_, _ = w.Write([]byte("tail hunt"))
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "tail hunt" {
		t.Errorf("Close did not flush: %q", got)
	}
}

// syncBuffer serializes Writes so concurrent tests can use a bytes.Buffer.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func TestWriter_ConcurrentWrites(t *testing.T) {
	var buf syncBuffer
	w := NewWriter(&buf, NewSet([]string{"API_TOKEN=hunter2-xyz"}, nil, nil))
	const goroutines, repeats = 8, 200
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < repeats; i++ {
				if _, err := w.Write([]byte("secret hunter2-xyz done\n")); err != nil {
					t.Error(err)
					return
				}
			}
		}()
	}
	wg.Wait()
	_ = w.Flush()
	out := buf.String()
	if strings.Contains(out, "hunter2-xyz") {
		t.Fatalf("raw secret leaked under concurrency")
	}
	if got, want := strings.Count(out, Placeholder), goroutines*repeats; got != want {
		t.Errorf("masked %d occurrences, want %d", got, want)
	}
}

func TestWriter_EmptyWrite(t *testing.T) {
	w, buf := newTestWriter("API_TOKEN=hunter2-xyz")
	n, err := w.Write(nil)
	if err != nil || n != 0 {
		t.Errorf("Write(nil) = (%d, %v)", n, err)
	}
	if buf.Len() != 0 {
		t.Errorf("empty write emitted bytes")
	}
}
