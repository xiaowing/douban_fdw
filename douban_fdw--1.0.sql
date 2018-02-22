/* douban_fdw/douban_fdw--1.0.sql */

-- complain if script is sourced in psql, rather than via CREATE EXTENSION
\echo Use "CREATE EXTENSION douban_fdw" to load this douban. \quit

CREATE FUNCTION douban_fdw_handler()
RETURNS fdw_handler
AS 'MODULE_PATHNAME'
LANGUAGE C STRICT;

CREATE FUNCTION douban_fdw_validator(text[], oid)
RETURNS void
AS 'MODULE_PATHNAME'
LANGUAGE C STRICT;

CREATE FOREIGN DATA WRAPPER douban_fdw
  HANDLER douban_fdw_handler
  VALIDATOR douban_fdw_validator;
