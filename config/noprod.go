//go:build !RELEASE
// +build !RELEASE

package config

// All the code inside !isProd branches will be removed by the compiler in
// release builds.
const IsProd = false
