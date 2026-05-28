//go:build js

package auth

import "github.com/menor/sol/internal/errors"

// SystemBrowser is a stub for browser builds — Sol-WASM never opens browsers.
type SystemBrowser struct{}

var _ BrowserOpener = (*SystemBrowser)(nil)

func (b *SystemBrowser) Open(url string) error {
	return errors.NewUnsupportedError("browser open not available in browser runtime")
}

// openBrowser is referenced by oauth.go (shared file); kept here as a stub
// so the package compiles under js. It is unreachable in the wasm path
// because auth:login is not exposed via the wasm entrypoint.
func openBrowser(url string) error {
	return errors.NewUnsupportedError("browser open not available in browser runtime")
}
