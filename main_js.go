//go:build js

package main

// The native CLI entrypoint is excluded from js/wasm builds.
// The wasm module is built from cmd/sol-wasm/main.go.
func main() {}
