package cmd

import "testing"

func TestParseSSHURL(t *testing.T) {
	tests := []struct {
		name     string
		sshURL   string
		wantArgs []string
	}{
		{
			name:     "simple user@host",
			sshURL:   "user@host.example.com",
			wantArgs: []string{"user@host.example.com"},
		},
		{
			name:     "with ssh:// prefix",
			sshURL:   "ssh://user@host.example.com",
			wantArgs: []string{"user@host.example.com"},
		},
		{
			name:     "with port",
			sshURL:   "user@host.example.com:2222",
			wantArgs: []string{"-p", "2222", "user@host.example.com"},
		},
		{
			name:     "with ssh:// prefix and port",
			sshURL:   "ssh://user@host.example.com:2222",
			wantArgs: []string{"-p", "2222", "user@host.example.com"},
		},
		{
			name:     "complex user with project ID",
			sshURL:   "ssh://abcd1234-main-abc123--app@ssh.us-3.platform.sh:22",
			wantArgs: []string{"-p", "22", "abcd1234-main-abc123--app@ssh.us-3.platform.sh"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSSHURL(tt.sshURL)
			if len(got) != len(tt.wantArgs) {
				t.Errorf("parseSSHURL(%q) = %v, want %v", tt.sshURL, got, tt.wantArgs)
				return
			}
			for i, arg := range got {
				if arg != tt.wantArgs[i] {
					t.Errorf("parseSSHURL(%q)[%d] = %q, want %q", tt.sshURL, i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestModifySSHURLForApp(t *testing.T) {
	tests := []struct {
		name    string
		sshURL  string
		appName string
		want    string
	}{
		{
			name:    "simple URL",
			sshURL:  "ssh://user@host.example.com",
			appName: "myapp",
			want:    "ssh://user--myapp@host.example.com",
		},
		{
			name:    "with port",
			sshURL:  "ssh://abcd1234-main-abc123@ssh.us-3.platform.sh:22",
			appName: "api",
			want:    "ssh://abcd1234-main-abc123--api@ssh.us-3.platform.sh:22",
		},
		{
			name:    "no @ symbol",
			sshURL:  "ssh://hostname",
			appName: "app",
			want:    "ssh://hostname",
		},
		{
			name:    "without ssh:// prefix",
			sshURL:  "user@host.example.com",
			appName: "worker",
			want:    "ssh://user--worker@host.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := modifySSHURLForApp(tt.sshURL, tt.appName)
			if got != tt.want {
				t.Errorf("modifySSHURLForApp(%q, %q) = %q, want %q", tt.sshURL, tt.appName, got, tt.want)
			}
		})
	}
}

func TestValidAppName(t *testing.T) {
	tests := []struct {
		name  string
		app   string
		valid bool
	}{
		{"valid simple", "myapp", true},
		{"valid with numbers", "app123", true},
		{"valid with hyphen", "my-app", true},
		{"valid with underscore", "my_app", true},
		{"valid complex", "App_Name-123", true},
		{"starts with number", "1app", true},
		{"starts with hyphen", "-app", false},
		{"starts with underscore", "_app", false},
		{"contains space", "my app", false},
		{"empty string", "", false},
		{"contains special char", "my@app", false},
		{"ssh injection attempt", "-oProxyCommand=bad", false},
		{"path traversal", "../etc/passwd", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validAppName.MatchString(tt.app)
			if got != tt.valid {
				t.Errorf("validAppName.MatchString(%q) = %v, want %v", tt.app, got, tt.valid)
			}
		})
	}
}
