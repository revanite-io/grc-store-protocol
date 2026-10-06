module github.com/revanite-io/grc-store-protocol

// Minimum Go: the code itself needs only go 1.18 (strings.Cut). The floor
// tracks the OLDEST Go release still getting security fixes (Go supports the
// two newest), not an author's machine version, so consumers aren't pushed
// past a supported toolchain. Raise it when that release goes out of support;
// CI tests the floor and the newest release.
go 1.26
