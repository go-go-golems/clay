package sql

import "testing"

func TestNormalizeDuckDBDSN(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
		want string
	}{
		{
			name: "absolute path",
			dsn:  "duckdb:///tmp/app.db",
			want: "/tmp/app.db",
		},
		{
			name: "absolute path with query params",
			dsn:  "duckdb:///tmp/app.db?access_mode=read_only&threads=4",
			want: "/tmp/app.db?access_mode=read_only&threads=4",
		},
		{
			name: "host-only relative path",
			dsn:  "duckdb://relative.db",
			want: "relative.db",
		},
		{
			name: "memory path",
			dsn:  "duckdb:///:memory:",
			want: ":memory:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeDuckDBDSN(tt.dsn)
			if err != nil {
				t.Fatalf("normalizeDuckDBDSN(%q) returned error: %v", tt.dsn, err)
			}
			if got != tt.want {
				t.Fatalf("normalizeDuckDBDSN(%q) = %q, want %q", tt.dsn, got, tt.want)
			}
		})
	}
}
