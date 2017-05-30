# Ewok

Ewok is a minimalistic RSS aggregator.

### Requirements
* A postgres server (tested on 9.6.3)

### Installation
1. Create a new database user called `rss`.
2. Run the SQL scripts in the `sql/` folder.
3. Make a config file similar to the provided example, and fill it with
   the db details.
4. Build the binaries (i.e. ewok and sync_feeds) with `go build`.
5. Make `sync_feeds` run periodically.
6. Setup a reverse proxy that points to ewok.
7. Add feeds by directly inserting into the db.
