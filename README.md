# go-hsr-oauth

Reusable HSR OAuth feature module.

## Layout

- `pkg/oauth`: public library API for other projects.
- `internal/repository`, `internal/service`, `internal/handler`: private OAuth implementation.
- User identity and session creation are delegated to `github.com/lwmacct/260630-go-hsr-auth`.
- Shared database, request context, token, and schema helpers live in `github.com/lwmacct/260630-go-hsr-shared`.

## Checks

```bash
go test -count=1 ./...
go test -count=1 ./internal/testutil/tddcheck
```
