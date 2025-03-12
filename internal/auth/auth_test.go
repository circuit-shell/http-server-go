package auth

import (
	"golang.org/x/crypto/bcrypt"
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	// Previous TestHashPassword implementation remains the same
	tests := []struct {
		name          string
		password      string
		wantErr       bool
		errorContains string
	}{
		{
			name:          "valid password",
			password:      "mySecurePassword123",
			wantErr:       false,
			errorContains: "",
		},
		{
			name:          "empty password",
			password:      "",
			wantErr:       false,
			errorContains: "",
		},
		{
			name:          "very long password",
			password:      strings.Repeat("a", 72), // bcrypt has a max length of 72 bytes
			wantErr:       false,
			errorContains: "",
		},
		
		{
			name:          "password exceeding max length",
			password:      strings.Repeat("a", 73), // should still work but will be truncated
			wantErr: true,
			errorContains: "password length exceeds 72 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := HashPassword(tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.errorContains) {
				t.Errorf("HashPassword() error = %v, should contain %v", err, tt.errorContains)
				return
			}

			if !tt.wantErr {
				if hashedPassword == "" {
					t.Error("HashPassword() returned empty hash")
					return
				}

				err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(tt.password))
				if err != nil {
					t.Errorf("Failed to verify hashed password: %v", err)
					return
				}

				if hashedPassword == tt.password {
					t.Error("HashPassword() returned unhashed password")
					return
				}
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "correct password",
			password: "mySecurePassword123",
			want:     true,
		},
		{
			name:     "incorrect password",
			password: "wrongPassword123",
			want:     false,
		},
		{
			name:     "empty password with empty hash",
			password: "",
			hash:     "",
			want:     false,
		},
		{
			name:     "empty password with valid hash",
			password: "",
			want:     false,
		},
		{
			name:     "invalid hash format",
			password: "mySecurePassword123",
			hash:     "invalid_hash_format",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// If hash is not provided in the test case, generate one
			var hash string
			if tt.hash == "" && tt.want {
				var err error
				hash, err = HashPassword(tt.password)
				if err != nil {
					t.Fatalf("Failed to generate hash for test: %v", err)
				}
			} else {
				hash = tt.hash
			}

			// Test the password check
			got := CheckPasswordHash(tt.password, hash)
			if got != tt.want {
				t.Errorf("CheckPasswordHash() = %v, want %v", got, tt.want)
			}

			// Additional verification for positive cases
			if tt.want {
				// Double-check with bcrypt directly
				err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(tt.password))
				if err != nil {
					t.Errorf("bcrypt.CompareHashAndPassword() failed but CheckPasswordHash() passed: %v", err)
				}
			}
		})
	}
}

// Benchmarks
func BenchmarkHashPassword(b *testing.B) {
	password := "mySecurePassword123"
	for i := 0; i < b.N; i++ {
		_, err := HashPassword(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCheckPasswordHash(b *testing.B) {
	password := "mySecurePassword123"
	hash, err := HashPassword(password)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckPasswordHash(password, hash)
	}
}
