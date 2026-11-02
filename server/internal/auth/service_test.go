package auth_test

import (
	"tano/internal/auth"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "VeryStrongPassword@1234",
			wantErr:  false,
		},
		{
			name:     "missing symbol",
			password: "P1a2S3s4",
			wantErr:  true,
		},
		{
			name:     "missing uppercase",
			password: "!holypassword69",
			wantErr:  true,
		},
		{
			name:     "too short",
			password: "H@1yPwd",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := auth.ValidatePassword(tt.password)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ValidatePassword() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ValidatePassword() succeeded unexpectedly")
			}
		})
	}
}
