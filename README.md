# ex-go

A minimal Go service (the platform's own "hello" dogfood fixture): it echoes
platform identity, verifies the `DEPLOYMIND_IDENTITY_KEY` header, proves
`DATABASE_URL` works, and answers `GET /` and `GET /healthz` with 200. It
declares `GREETING_SECRET` and `primitives.db.enabled: true`.

Deploy it (deploymind.yaml is already committed, so `init` is not needed):

```
deploymind validate && git push && deploymind plan && deploymind deploy
```
