package ui

import "testing"

func TestFormatDataSize(t *testing.T) {
	cases := []struct {
		bytes int64
		want  string
	}{
		{0, "0 Б"}, {1, "1 Б"}, {1023, "1023 Б"},
		{1024, "1,00 КиБ"}, {1343886, "1,28 МиБ"},
		{1048576, "1,00 МиБ"}, {1073741824, "1,00 ГиБ"}, {-1, "0 Б"},
	}
	for _, tc := range cases {
		if got := formatDataSize(tc.bytes); got != tc.want {
			t.Errorf("formatDataSize(%d) = %q, want %q", tc.bytes, got, tc.want)
		}
	}
}
