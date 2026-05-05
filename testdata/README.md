# crap4go testdata

This directory contains fixture files used by the test suites across
`internal/complexity`, `internal/coverage`, `internal/config`, and `internal/pipeline`.

## Structure

```
testdata/
├── complexity/        # .go source fixtures for CC extraction tests
│   ├── simple.go      # single function, base CC
│   ├── branches.go    # if/for/switch
│   ├── logical.go     # &&/|| operators (incl. non-condition context pinning)
│   ├── methods.go     # methods + pointer receivers
│   ├── closures.go    # nested FuncLit
│   └── initfuncs.go   # multiple init() — deduplication test
├── coverage/          # coverage.out fixtures for parser tests
│   ├── simple.out     # minimal "set" mode profile
│   ├── count.out      # "count" mode profile
│   ├── atomic.out     # "atomic" mode profile
│   └── closure.out    # outer+inner closure — innermost attribution test
├── config/            # config file fixtures for loader tests
│   ├── minimal.json   # {"threshold": 20} — only threshold set
│   ├── minimal.toml   # threshold = 20 — only threshold set
│   ├── full.json      # all supported fields (coverprofile, threshold, all, paths)
│   ├── full.toml      # same as full.json in TOML syntax
│   ├── invalid.json   # malformed JSON — triggers parse error
│   ├── invalid.toml   # malformed TOML — triggers parse error
│   └── config.yaml    # unsupported extension — triggers extension error
└── pipeline/          # integration fixture Go packages
    ├── foo/           # valid package with tests + coverage profile
    ├── bad/           # package with invalid Go syntax (parse error test)
    ├── empty/         # empty directory (no .go files)
    └── buildtag/      # package with //go:build ignore (build-tag skip test)
```
