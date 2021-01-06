# multihealthcheck

`multihealthcheck` is a health check aggregator.

When `multihealthcheck` receives an http request, it makes parallel requests in
turn to the configured http/https health check targets. If all the defined
health checks are OK, it returns a 200 OK. Otherwise, it returns a 500 DOWN.

## Status: **WIP**

All areas of the code and this daemon's behavior are subject to breaking
changes, potentially without any mention in commit messages or release notes.

### Areas needing improvement

*   Documentation
*   Better config (this will probably be breaking)
*   Concurrency limiting and healthiness caching
*   Test coverage

## Contributing

Contributions considered, but be aware that this is mostly just something we
needed. It's public because there's no reason anyone else should have to waste
an afternoon building something similar.

This project is licensed under the [Apache License, Version 2.0](LICENSE).

Please include a `Signed-off-by` in all commits, per
[Developer Certificate of Origin version 1.1](DCO).
