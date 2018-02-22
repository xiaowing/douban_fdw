/*-------------------------------------------------------------------------
 *
 * douban_fdw.c
 * HelloWorld of foreign-data wrapper.
 *
 * written by Wataru Ikarashi <wikrsh@gmail.com>
 *
 *-------------------------------------------------------------------------
 */

#include <sys/stat.h>
#include <unistd.h>

#include "libdbango.h"

PG_MODULE_MAGIC;

/*
 * SQL functions
 */
extern Datum douban_fdw_handler(PG_FUNCTION_ARGS);
extern Datum douban_fdw_validator(PG_FUNCTION_ARGS);

PG_FUNCTION_INFO_V1(douban_fdw_handler);
PG_FUNCTION_INFO_V1(douban_fdw_validator);

/*
 * Foreign-data wrapper handler function
 */
Datum
douban_fdw_handler(PG_FUNCTION_ARGS)
{
    FdwRoutine *fdwroutine = makeNode(FdwRoutine);
    
    fdwroutine->GetForeignRelSize = doubanGetForeignRelSize_cgo;
    fdwroutine->GetForeignPaths = doubanGetForeignPaths_cgo;
    fdwroutine->GetForeignPlan = doubanGetForeignPlan_cgo;
    fdwroutine->ExplainForeignScan = doubanExplainForeignScan_cgo;
    fdwroutine->BeginForeignScan = doubanBeginForeignScan_cgo;
    fdwroutine->IterateForeignScan = doubanIterateForeignScan_cgo;
    fdwroutine->ReScanForeignScan = doubanReScanForeignScan_cgo;
    fdwroutine->EndForeignScan = doubanEndForeignScan_cgo;
    fdwroutine->AnalyzeForeignTable = doubanAnalyzeForeignTable_cgo;

    PG_RETURN_POINTER(fdwroutine);
}

/*
 * Validate the generic options given to a FOREIGN DATA WRAPPER, SERVER
 * USER MAPPING or FOREIGN TABLE that uses douban_fdw.
 */
Datum
douban_fdw_validator(PG_FUNCTION_ARGS)
{
    List		*options_list = untransformRelOptions(PG_GETARG_DATUM(0));
    Oid			catalog = PG_GETARG_OID(1);
    ListCell	*cell;
    int         kind = 0;
  
    foreach(cell, options_list)
    {
        DefElem	 *def = (DefElem *) lfirst(cell);
        kind = checkOptionName(def->defname, catalog);

        switch (kind) 
        {
            case 0:    /* "rank_name" */
                if (!checkRankName(defGetString(def)))
                {
                    ereport(ERROR, 
                        (errcode(ERRCODE_FDW_TABLE_NOT_FOUND), 
                        errmsg("specified rank name \"%s\" does not exist", defGetString(def))));
                }
                break;
            default:
                ereport(ERROR, 
                    (errcode(ERRCODE_FDW_INVALID_OPTION_NAME), 
                    errmsg("invalid option \"%s\" for the object %d", def->defname, catalog)));
        }
    }

    PG_RETURN_VOID();
}