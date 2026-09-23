package engine

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Cunyule/AI-coding/internal/model"
)

func TestIdentify(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	e, err := Load(filepath.Join(filepath.Dir(filename), "..", "..", "rules", "fingerprints.json"))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, banner, protocol, product, version, os string
		port                                         int
	}{
		{"openssh", "SSH-2.0-OpenSSH_8.9p1 Ubuntu-3", "SSH", "OpenSSH", "8.9p1", "Ubuntu", 22},
		{"nginx nonstandard port", "HTTP/1.1 200 OK\r\nServer: nginx/1.25.3", "HTTP", "nginx", "1.25.3", "", 8443},
		{"apache", "HTTP/1.1 200 OK\r\nServer: Apache/2.4.57", "HTTP", "Apache", "2.4.57", "", 443},
		{"mysql", "J\x00\x00\x00\n8.0.32\x00", "MySQL", "MySQL", "8.0.32", "", 3306},
		{"redis", "+PONG", "Redis", "Redis", "", "", 6379},
		{"proftpd", "220 ProFTPD 1.3.7 Server", "FTP", "ProFTPD", "1.3.7", "", 21},
		{"unknown", "QUIT\r\n", "unknown", "", "", "", 12345},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.Identify(model.ScanRecord{IP: "1.2.3.4", Port: tt.port, Banner: tt.banner})
			if got.Protocol != tt.protocol || got.Product != tt.product || got.Version != tt.version || got.OSHint != tt.os {
				t.Fatalf("got %+v", got)
			}
		})
	}
}
