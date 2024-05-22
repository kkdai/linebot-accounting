package main

import (
	"context"
	"os"
	"testing"
)

func TestFuncCall(t *testing.T) {
	gap := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	firebaseURL := os.Getenv("FIREBASE_URL")
	geminiKey = os.Getenv("GOOGLE_GEMINI_API_KEY")

	// If no environment variable, goskip the test
	if gap == "" || firebaseURL == "" {
		t.Skip("No environment variable")
	}
	ctx := context.Background()
	initFirebase(gap, firebaseURL, ctx)
	fireDB.SetPath("accounting/u1234")
	gemini.GeminiFunctionCall("Can you record an expense of 50 dollars for workout out on 2024-05-17?")
}
