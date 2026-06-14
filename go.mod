module github.com/revanite-io/grc-store-protocol

// Minimum Go: the code uses only strings.Cut (go 1.18). The floor is set to 1.22
// for toolchain hygiene — deliberately NOT an author's machine version, so
// external consumers aren't forced onto a newer toolchain. Do not bump without a
// real language-feature reason.
go 1.22
