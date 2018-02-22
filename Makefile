# douban_fdw/Makefile

MODULE_big = douban_fdw
OBJS = douban_fdw.o

EXTENSION = douban_fdw
DATA = douban_fdw--1.0.sql

REGRESS = douban_fdw

CC = gcc

GOC = go
GOFLAG = -buildmode=c-shared
LIBNAME = dbango
GOLIBNAME = lib${LIBNAME}
GOLIB = ${GOLIBNAME}.so
GOHEADER = ${GOLIBNAME}.h
GOSRC := $(wildcard *.go)

SHLIB_LINK = -L./ -ldbango
 
export CGO_CFLAGS = -I$(shell $(PG_CONFIG) --includedir-server)

all: all-golib 

all-golib: ${GOLIB}
${GOLIB}: ${GOSRC}
	${GOC} build -o ${GOLIB} ${GOFLAG} ${GOSRC}

${OBJS}: ${GOLIB}

clean: clean-golib
clean-golib:
	-rm -f ${GOLIB} ${GOHEADER}

install: install-golib
install-golib:
	-$(INSTALL_SHLIB) ${GOLIB} '$(DESTDIR)$(pkglibdir)/${GOLIB}'

uninstall: uninstall-golib
uninstall-golib:
	-rm -f '$(DESTDIR)$(pkglibdir)/${GOLIB}'

PG_CONFIG = pg_config
PGXS := $(shell $(PG_CONFIG) --pgxs)
include $(PGXS)
