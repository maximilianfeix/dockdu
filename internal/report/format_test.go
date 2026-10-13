package report

import "testing"

func TestBytes(t *testing.T) {
	cases := map[int64]string{
		-1:           "unknown",
		0:            "0 B",
		1023:         "1023 B",
		1024:         "1.0 KiB",
		1536:         "1.5 KiB",
		6656000000:   "6.2 GiB",
		245106737152: "228.3 GiB",
		1 << 40:      "1.0 TiB",
	}
	for in, want := range cases {
		if got := Bytes(in); got != want {
			t.Errorf("Bytes(%d) = %q, want %q", in, got, want)
		}
	}
}
