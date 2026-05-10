package daemon

import "testing"

func TestCommandLooksLikeDaemon(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    bool
	}{
		{
			name:    "installed foreground daemon",
			command: "/Users/alice/.local/bin/clipterm daemon --foreground",
			want:    true,
		},
		{
			name:    "debug foreground daemon",
			command: "/Users/alice/.local/bin/clipterm daemon --foreground --debug-hotkeys",
			want:    true,
		},
		{
			name:    "different clipterm command",
			command: "/Users/alice/.local/bin/clipterm daemon --status",
			want:    false,
		},
		{
			name:    "different process",
			command: "/usr/bin/vim daemon --foreground",
			want:    false,
		},
		{
			name:    "empty command",
			command: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := commandLooksLikeDaemon(tt.command); got != tt.want {
				t.Fatalf("commandLooksLikeDaemon(%q) = %v, want %v", tt.command, got, tt.want)
			}
		})
	}
}
