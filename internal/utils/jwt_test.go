package utils

import "testing"

func TestGenerateAndValidateAccessToken(t *testing.T) {
	token, err := GenerateAccessToken(1, "tester")
	if err != nil {
		t.Fatalf("generate access token failed: %v", err)
	}
	claims, err := ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("validate access token failed: %v", err)
	}
	if claims.UserID != 1 || claims.Username != "tester" || claims.TokenType != TokenTypeAccess {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestGenerateAndValidateRefreshToken(t *testing.T) {
	token, claims, err := GenerateRefreshToken(2, "refresh_user")
	if err != nil {
		t.Fatalf("generate refresh token failed: %v", err)
	}
	if claims.TokenType != TokenTypeRefresh {
		t.Fatalf("unexpected token type: %s", claims.TokenType)
	}
	if _, err = ValidateRefreshToken(token); err != nil {
		t.Fatalf("validate refresh token failed: %v", err)
	}
}
