// Copyright 2026 José Menor
// Licensed under the Apache License, Version 2.0.
// See LICENSE and NOTICE files for details.

// Package auth implements OAuth2 + PKCE authentication against Upsun's auth
// server (auth.upsun.com). It is split across two files with distinct
// responsibilities:
//
// # service.go — interactive CLI operations
//
// Service orchestrates the login flow, logout, and status checks. Use it when
// your code drives user interaction:
//
//	svc := auth.DefaultService()
//	result, err := svc.Login(ctx, auth.LoginOptions{OnProgress: ...})
//	err = svc.Logout(ctx)
//	status, err := svc.Status(ctx)
//
// Service accepts context per method call, following standard Go conventions.
// Its dependencies (TokenStore, BrowserOpener) are injected, making it
// straightforward to test without touching the OS keychain or a real browser.
//
// # token.go — API client integration
//
// TokenSource returns an oauth2.TokenSource for injecting tokens into HTTP
// requests. Use it when building an API client that needs automatic token
// refresh:
//
//	ts, err := auth.TokenSource(ctx)
//	httpClient := oauth2.NewClient(ctx, ts)
//
// TokenSource checks UPSUN_TOKEN first (CI/automated environments), then falls
// back to the keychain token with automatic refresh on expiry.
//
// Why context is stored in keyringTokenSource: the oauth2.TokenSource interface
// defines Token() with no context parameter. Storing the context in the struct
// is the only way to pass it through to the token refresh call. This is a
// documented exception to Go's "don't store context in structs" guideline. For
// long-lived token sources, create a new one with a fresh context rather than
// reusing a stale one.
//
// # Choosing between Service and TokenSource
//
//   - CLI commands that log in, log out, or show auth status → Service
//   - Creating an HTTP client for API calls → TokenSource
package auth
