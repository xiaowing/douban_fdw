package main

/*
#include "postgres.h"

#include "access/htup_details.h"
#include "access/reloptions.h"
#include "access/sysattr.h"
#include "catalog/pg_foreign_table.h"
#include "commands/copy.h"
#include "commands/defrem.h"
#include "commands/explain.h"
#include "commands/vacuum.h"
#include "foreign/fdwapi.h"
#include "foreign/foreign.h"
#include "funcapi.h"
#include "miscadmin.h"
#include "nodes/makefuncs.h"
#include "optimizer/cost.h"
#include "optimizer/pathnode.h"
#include "optimizer/planmain.h"
#include "optimizer/restrictinfo.h"
#include "optimizer/var.h"
#include "utils/memutils.h"
#include "utils/rel.h"

// Definitions of gateway functions
void doubanGetForeignRelSize_cgo(PlannerInfo *root,
                       RelOptInfo *baserel,
					   Oid foreigntableid)
{
	void doubanGetForeignRelSize(PlannerInfo *, RelOptInfo *, Oid);
	doubanGetForeignRelSize(root, baserel, foreigntableid);
}

void doubanGetForeignPaths_cgo(PlannerInfo *root,
                       RelOptInfo *baserel,
					   Oid foreigntableid)
{
	void doubanGetForeignPaths(PlannerInfo *, RelOptInfo *, Oid);
	doubanGetForeignPaths(root, baserel, foreigntableid);
}

ForeignScan *doubanGetForeignPlan_cgo(PlannerInfo *root,
                    RelOptInfo *baserel,
                    Oid foreigntableid,
                    ForeignPath *best_path,
                    List *tlist,
                    List *scan_clauses,
					Plan *outer_plan)
{
	ForeignScan *doubanGetForeignPlan(PlannerInfo *, RelOptInfo *,
		Oid, ForeignPath *, List *, List *, Plan *);
	return doubanGetForeignPlan(root, baserel, foreigntableid, best_path,
		tlist, scan_clauses, outer_plan);
}

void doubanExplainForeignScan_cgo(ForeignScanState *node,
						ExplainState *es)
{
	void doubanExplainForeignScan(ForeignScanState *, ExplainState *);
	doubanExplainForeignScan(node, es);
}

void doubanBeginForeignScan_cgo(ForeignScanState *node,
					  int eflags)
{
	void doubanBeginForeignScan(ForeignScanState *, int);
	doubanBeginForeignScan(node, eflags);
}

TupleTableSlot *doubanIterateForeignScan_cgo(ForeignScanState *node)
{
	TupleTableSlot *doubanIterateForeignScan(ForeignScanState *);
	return doubanIterateForeignScan(node);
}

void doubanReScanForeignScan_cgo(ForeignScanState *node)
{
	void doubanReScanForeignScan(ForeignScanState *);
	doubanReScanForeignScan(node);
}

void doubanEndForeignScan_cgo(ForeignScanState *node)
{
	void doubanEndForeignScan(ForeignScanState *);
	doubanEndForeignScan(node);
}

bool doubanAnalyzeForeignTable_cgo(Relation relation,
                         AcquireSampleRowsFunc *func,
						 BlockNumber *totalpages)
{
	bool doubanAnalyzeForeignTable(Relation,
		AcquireSampleRowsFunc *, BlockNumber *);
	return doubanAnalyzeForeignTable(relation, func, totalpages);
}
*/
import "C"
