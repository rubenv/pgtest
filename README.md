# pgtest

> Go library to spawn single-use PostgreSQL servers for unit testing

[![Build Status](https://github.com/rubenv/pgtest/workflows/Test/badge.svg)](https://github.com/rubenv/pgtest/actions) [![GoDoc](https://godoc.org/github.com/rubenv/pgtest?status.png)](https://godoc.org/github.com/rubenv/pgtest)

Spawns a PostgreSQL server with a single database configured. Ideal for unit
tests where you want a clean instance each time. Then clean up afterwards.

Features:

* Starts a clean isolated PostgreSQL database
* Tested on Fedora, Ubuntu and Alpine
* Optimized for in-memory execution, to speed up unit tests
* Less than 1 second startup / initialization time

## Usage

In your unit test:
```go
pg, err := pgtest.Start()
defer pg.Stop()

// Do something with pg.DB (which is a *sql.DB)
```

## License

This library is distributed under the [MIT](LICENSE) license.
