package parser

import "testing"

// FuzzParser asserts the parser never panics on arbitrary input and always
// returns a non-nil File.
func FuzzParser(f *testing.F) {
	seeds := []string{
		"",
		"set default := \"greet\"\n",
		"greet name=\"world\":\n    @echo hi {{name}}\n",
		"build: greet\n    echo build\n",
		"[cache(inputs=[\"a\"], outputs=[\"b\"])]\nx:\n    echo x\n",
		"a: (b \"arg\") c && d\n    echo a\n",
		"x := if a == \"1\" { \"y\" } else { \"z\" }\n",
		"deploy: docker::push\n    echo done\n",
		"mod sub \"sub.rune\"\nimport? \"opt.rune\"\n",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, src string) {
		file, _ := Parse("fuzz", src)
		if file == nil {
			t.Fatal("Parse returned nil File")
		}
	})
}
