# douban_fdw
A PostgreSQL's Foreign Data Wrapper (FDW) for retrieving the movie ranking data via the public API of douban.com. This FDW is mainly written in GO(cgo).

This toy FDW was inspired by the github repositories as follows

* [rapidloop/ptgo](https://github.com/rapidloop/ptgo)
* [umitanuki/twitter_fdw](https://github.com/umitanuki/twitter_fdw)

## Usage

### how to install from source

1. make sure the path to the PostgreSQL's binary had been exported to the `PATH` environment variable because it is necessary to use the `pg_config` command for the source build. 

    please use `which pg_config` to check the path

2. make sure the GO language was successfully installed and the path to `go` had been exported to the `PATH` environment variable

    please use `go version` to check that

3. make sure the toolchain of `gcc` and `make` were successfully installed

4. execute the following command

    ````sh
    $cd /path/to/douban_fdw
    $make
    $make install
    ````

    if the commands above successed, the shared library files would be installed into `/path/to/postgres/install/lib` and the other files would be installed into `/path/to/postgres/install/share`

5. restart the postgres instance

6. use the postgres client (such as `psql`) to connect to the postgres instance and execute the following SQL statement and it should be done by the superuser

    ````sql
    CREATE EXTENSION douban_fdw;
    ````

### how to use the FDW

0. make sure that the postgresql server where the fdw was installed to is able to access *douban.com* on internet

1. first of all, create a foreign server like

    ````sql
    CREATE SERVER {servername} FOREIGN DATA WRAPPER douban_fdw;
    ````

2. define a foreign table with the foreign server above

    ````sql
    CREATE FOREIGN TABLE {tablename} (rating {data type}...) SERVER {servername} OPTIONS (rank_name 'top250');
    ````

    *you can name the foreign table whatever you wanted to, but the column name should be as follows. if you defined an column name out of the valid range, it would cause an error when you queried the table*

    *the column names which can be identified:*

    * casts
    * collectcount
    * directors
    * genres
    * id
    * originname
    * rating
    * subtype
    * title
    * url
    * year

3. use the SELECT statement to query the foreign table defined above

## Limitations

1. this FDW currently can work with **PostgreSQL 9.5** only, because the internal fdw interfaces changed

2. the foreign table defined by douban_fdw can work properly only on the database of which the encoding being **UTF8**, because most of the data retrieved from douban.com are simplified chinese characters (encoded in UTF8)

3. according to the official manual, **the douban's public api can only be called within 40 times per-hour from one ip address**, currently the user can only query the foreign table less than 40 time in a hour

4. it only supports the public movie api of "top250"(*/v2/movie/top250*) currently

## TODO

the following features are on the way

- [x] the implementation of the rescan routine
- [ ] a local persistant buffer to solve the times limit issue of the douban API
- [ ] support the public api for retrieving data of chart "us_box"
- [ ] server-side encoding convert to support the database of which not being UTF8-encoded
- [ ] PostgreSQL 9.6+ support
