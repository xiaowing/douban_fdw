package main

/*
#include "postgres.h"

#cgo LDFLAGS: -Wl,--unresolved-symbols=ignore-all

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
#include "nodes/pg_list.h"
#include "optimizer/cost.h"
#include "optimizer/pathnode.h"
#include "optimizer/planmain.h"
#include "optimizer/restrictinfo.h"
#include "optimizer/var.h"
#include "utils/memutils.h"
#include "utils/rel.h"
#include "utils/syscache.h"

static void ErrorReport(int sqlstate, char *error_msg)
{
	ereport(ERROR, (errcode(sqlstate), errmsg(error_msg)));
}

static void InfoReport(char *error_msg)
{
	ereport(INFO, (errmsg(error_msg)));
}
*/
import "C"
import (
	"strconv"
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

const (
	EnumRankName = iota
)

// the following constants are calculated base on
// the definitions of the referred SQLState defined in "utils/errcodes.h"
const (
	ERRCODE_FDW_ERROR                                  = 2456
	ERRCODE_FDW_COLUMN_NAME_NOT_FOUND                  = 83888536
	ERRCODE_FDW_DYNAMIC_PARAMETER_VALUE_NEEDED         = 33556888
	ERRCODE_FDW_FUNCTION_SEQUENCE_ERROR                = 264600
	ERRCODE_FDW_INCONSISTENT_DESCRIPTOR_INFORMATION    = 17303960
	ERRCODE_FDW_INVALID_ATTRIBUTE_VALUE                = 67635608
	ERRCODE_FDW_INVALID_COLUMN_NAME                    = 117442968
	ERRCODE_FDW_INVALID_COLUMN_NUMBER                  = 134220184
	ERRCODE_FDW_INVALID_DATA_TYPE                      = 67111320
	ERRCODE_FDW_INVALID_DATA_TYPE_DESCRIPTORS          = 100665752
	ERRCODE_FDW_INVALID_DESCRIPTOR_FIELD_IDENTIFIER    = 19138968
	ERRCODE_FDW_INVALID_HANDLE                         = 301992344
	ERRCODE_FDW_INVALID_OPTION_INDEX                   = 318769560
	ERRCODE_FDW_INVALID_OPTION_NAME                    = 335546776
	ERRCODE_FDW_INVALID_STRING_LENGTH_OR_BUFFER_LENGTH = 2361752
	ERRCODE_FDW_INVALID_STRING_FORMAT                  = 285215128
	ERRCODE_FDW_INVALID_USE_OF_NULL_POINTER            = 150997400
	ERRCODE_FDW_TOO_MANY_HANDLES                       = 67373464
	ERRCODE_FDW_NO_SCHEMAS                             = 536873368
	ERRCODE_FDW_OPTION_NAME_NOT_FOUND                  = 436210072
	ERRCODE_FDW_REPLY_HANDLE                           = 452987288
	ERRCODE_FDW_SCHEMA_NOT_FOUND                       = 553650584
	ERRCODE_FDW_TABLE_NOT_FOUND                        = 570427800
	ERRCODE_FDW_UNABLE_TO_CREATE_EXECUTION             = 469764504
	ERRCODE_FDW_UNABLE_TO_CREATE_REPLY                 = 486541720
	ERRCODE_FDW_UNABLE_TO_ESTABLISH_CONNECTION         = 503318936
)

/* 
 * the following const is used for retrieving data from fdw_private 
 * usage: pass the const as the 2nd argument of list_nth() while retrieving data
 */
const (
	FDW_PRIVATE_INDEX_ATTRUSED = 0
)

// A hashtable for the valid options name
var ValidOptions = map[string]DoubanOption{
	"rank_name": {"rank_name", C.ForeignTableRelationId, C.int(EnumRankName)},
}

type DoubanOption struct {
	name       string
	oid        C.Oid
	identifier C.int
}

// The state passed for the scan.
type DoubanScanState struct {
	scanningRel     *C.RelationData
	attrsRetrieved  []*TargetColumnMeta
	resultSet       []MovieItem
	currentRow      int
}

type TargetColumnMeta struct {
	attrNum       int        /* same as attrNum in PG, starts from 1 */
	attrName      string
	attrType      C.Oid
	attrTypmod    C.int32
}

func (self *TargetColumnMeta) convert2Datum(source *MovieItem) C.Datum {
	var typeinput C.regproc
	var typemod C.int
	var tuple *C.HeapTupleData
	var valueDatum, retvalDatum C.Datum

	tuple = C.SearchSysCache(C.TYPEOID, C.Datum(self.attrType), 
				C.Datum(0), C.Datum(0), C.Datum(0))
	if tuple == (*C.HeapTupleData)(nil) {
		ereport(ERRCODE_FDW_ERROR, "cache lookup failed for type%d", int(self.attrType))
	}

	pgtype := (*C.FormData_pg_type)(unsafe.Pointer(uintptr((unsafe.Pointer)(tuple.t_data)) + 
	    uintptr(((*C.HeapTupleHeaderData)((unsafe.Pointer)(tuple.t_data))).t_hoff)))
	typeinput = pgtype.typinput
	typemod   = C.int(pgtype.typtypmod)
	C.ReleaseSysCache(tuple)

	/* convert the field valut into Datum(de facto: CString) so that we can use the typinput function */
	tempMovieItem := *source
	movieItemType := reflect.TypeOf(tempMovieItem)
	movieItemValue := reflect.ValueOf(tempMovieItem)

	var itemValStr string
	switch self.attrName {
		case "rating":
			itemValStr = strconv.FormatFloat(float64(source.GetAverageScore()), 'f', 2, 32)
		case "genres":
			itemValStr = source.GetGenres()
		case "casts":
			itemValStr = source.GetCasts()
		case "directors":
			itemValStr = source.GetDirectors()
		case "collectcount":
			itemValStr = strconv.Itoa(source.CollectCount)
		default:
			for j := 0; j < movieItemType.NumField(); j++ {
				if self.attrName == strings.ToLower(movieItemType.Field(j).Name) {
					itemValStr = movieItemValue.FieldByName(movieItemType.Field(j).Name).String()
				}
			}
	}
	valueDatum = C.Datum(uintptr(unsafe.Pointer(C.CString(itemValStr))))

	retvalDatum = C.OidFunctionCall3Coll(C.Oid(typeinput), C.Oid(0), valueDatum, C.Datum(0), C.Datum(typemod))
	return retvalDatum
}

func ereport(sqlstate int, msgFormat string, args ...interface{}) {
	C.ErrorReport(C.int(sqlstate), C.CString(fmt.Sprintf(msgFormat, args...)))
}

func info(msgFormat string, args ...interface{}) {
	msg := C.CString(fmt.Sprintf(msgFormat, args...))
	C.InfoReport(msg)
	C.free(unsafe.Pointer(msg))
}

//export doubanGetForeignRelSize
func doubanGetForeignRelSize(root *C.PlannerInfo,
	baserel *C.RelOptInfo, foreigntableid C.Oid) {
	var referredAttrs *C.Bitmapset

	// Collect all the attributes needed for joins or final output.
	targetlist := (*C.Node)(unsafe.Pointer(baserel.reltargetlist)) // TODO: member field of 'RelOptInfo' changed in 9.6
	C.pull_varattnos(targetlist, baserel.relid, (**C.Bitmapset)(unsafe.Pointer(&referredAttrs)))

	// Add all the attributes used by restriction clauses.
	restrictNum := int(C.list_length(baserel.baserestrictinfo))
	for i := 0; i < restrictNum; i++ {
		rinfo := (*C.RestrictInfo)(unsafe.Pointer(uintptr(C.list_nth(baserel.baserestrictinfo, C.int(i)))))
		C.pull_varattnos((*C.Node)(unsafe.Pointer(rinfo.clause)), baserel.relid,
				(**C.Bitmapset)(unsafe.Pointer(&referredAttrs)))
	}
	
	// check if the name of the referred attrs are valid
	attributesRetrieved := referredFieldsValidator(foreigntableid, referredAttrs)
	C.bms_free(referredAttrs)

	baserel.fdw_private = Save(attributesRetrieved)
	baserel.rows = C.double(MovieRankingTop250Num)
	//TODO: width
}

// check whether the columns referred in a query are valid or not.
// and the phrase "valid" here has two meanings:
//   1. the columns appearred in the target list of the query are correct
//      i.e. the columns name are the ones defined in the CREATE FOREIGN TABLE
//   2. the referred columns do actully exist in the foreign data source
//
//   PostgreSQL's parser would garantee step 1 but cannot garantee step 2,
// that's why the following function is necessary
//
// Note:
//   if something goes wrong during the process of validation, it will directly
//   throw an error and long-jump just as the most functions in PostgreSQL did.
func referredFieldsValidator(foreigntableId C.Oid, referredFields *C.Bitmapset) []*TargetColumnMeta {
	var relation C.Relation
	var tupdesc C.TupleDesc

	// the magic number 1 means the lock type "AccessShareLock"
	// heap_open is an actual function. on the other hand, heap_close is a macro-function,
	// so we must use relation_close
	relation = C.heap_open(foreigntableId, 1)
	defer C.relation_close(relation, 1)

	tupdesc = (C.TupleDesc)(unsafe.Pointer(relation.rd_att))
	nattrs := int(tupdesc.natts)

	retval := make([]*TargetColumnMeta, 0, nattrs)

	tempMovieItem := MovieItem{}
	movieItemType := reflect.TypeOf(tempMovieItem)    /* a struct-type is necessary for calling the NumField() method */

	attrFound := false
	attrslice := (*[1 << 30]C.Form_pg_attribute)(unsafe.Pointer(tupdesc.attrs))[:nattrs:nattrs]

	for i := 1; i <= nattrs; i++ {
		attr := attrslice[i - 1]

		attrFound = false
		// Ignore dropped attributes.
		if attr.attisdropped == C.bool(1) {
			continue
		}

		// the current field not hit with the referredFields
		if C.bms_is_member(C.int(i-(-8)), referredFields) == C.bool(0) {
			continue
		}

		attrNameData := (*C.NameData)(unsafe.Pointer(&(attr.attname)))
		attrName := C.GoString(&(attrNameData.data[0]))

		for j := 0; j < movieItemType.NumField(); j++ {
			if strings.ToLower(attrName) == strings.ToLower(movieItemType.Field(j).Name) {
				if attrName == "imags" {
					ereport(ERRCODE_FDW_INVALID_COLUMN_NAME, "\"%s\" cannot be used as a column name", attrName)
				}
				attrFound = true
				break
			}
		}

		if !attrFound {
			ereport(ERRCODE_FDW_COLUMN_NAME_NOT_FOUND, "invalid column name \"%s\"", attrName)
		}

		// build the TargetColumnMeta for retval
		meta := new(TargetColumnMeta)
		meta.attrNum = i
		meta.attrName = attrName
		meta.attrType = attr.atttypid
		meta.attrTypmod = attr.atttypmod

		retval = append(retval, meta)
	}

	return retval
}

//export doubanGetForeignPaths
func doubanGetForeignPaths(root *C.PlannerInfo,
	baserel *C.RelOptInfo, foreigntableid C.Oid) {
	//TODO: improve the algorithm of cost estimate.
	startupCost := C.Cost(40.0)
	totalCost := C.Cost(40.0 + baserel.rows)

	path := (*C.Path)(unsafe.Pointer(C.create_foreignscan_path(root, baserel,
		baserel.rows, startupCost, totalCost, (*C.List)(nil),
		nil, nil, (*C.List)(nil))))
	C.add_path(baserel, path)
}

//export doubanGetForeignPlan
func doubanGetForeignPlan(root *C.PlannerInfo,
	baserel *C.RelOptInfo, foreigntableid C.Oid,
	bestPath *C.ForeignPath, tlist *C.List,
	scanClauses *C.List, outerPlan *C.Plan) *C.ForeignScan {

	_, ok := Restore(unsafe.Pointer(baserel.fdw_private)).([]*TargetColumnMeta)
	if (!ok) {
		ereport(ERRCODE_FDW_ERROR, "type assersion of \"%p\" to \"[]*TargetColumnMeta\" failed ", unsafe.Pointer(baserel.fdw_private))
	}
	//build the fdw_private list as the 5th arguement of make_foreignscan() if necessary
	scan_private := 
	    (*C.List)(unsafe.Pointer(C.lcons(unsafe.Pointer(baserel.fdw_private), (*C.List)(nil))))

	newScanClauses :=
		(*C.List)(unsafe.Pointer(C.extract_actual_clauses(scanClauses, C.bool(0))))

	result := (*C.ForeignScan)(unsafe.Pointer(C.make_foreignscan(tlist, newScanClauses, baserel.relid,
		(*C.List)(nil), scan_private,
		(*C.List)(nil), (*C.List)(nil), outerPlan)))

	return result
}

//export doubanBeginForeignScan
func doubanBeginForeignScan(node *C.ForeignScanState,
	eflags C.int) {
	var rel *C.RelationData

	// Do nothing in EXPLAIN (no ANALYZE) case.
	// macro "EXEC_FLAG_EXPLAIN_ONLY" means 0x0001
	if int(eflags) & 0x0001 != 0 {
		return
	}

	sstate := (*C.ScanState)(unsafe.Pointer(&(node.ss)))
	rel = (*C.RelationData)(unsafe.Pointer(sstate.ss_currentRelation))

	rank := getRankNameFromForeginTable(rel)

	items, err := RetrieveRankingData(rank, 50) //TODO: constant variable 50 should be changed
	if err != nil {
		ereport(ERRCODE_FDW_ERROR,
			"error occurred while retrieving data from douban.com\n  details:%v", err)
		return // ereport will cause the statement jumped out of the execution.
	}

	if (len(items) < MovieRankingTop250Num) {
		//TODO: change info into warning
		info("%d items expected from Douban.com, but only %d returned actually", MovieRankingTop250Num, len(items))
	}

	// get the fdw's private field from the ForeignScan
	pstate := (*C.PlanState)(unsafe.Pointer(&(sstate.ps)))
	fsplan := (*C.ForeignScan)(unsafe.Pointer(pstate.plan))
	privateList := (*C.List)(unsafe.Pointer(fsplan.fdw_private))
	fdwPrivate := unsafe.Pointer(uintptr(C.list_nth(privateList, C.int(FDW_PRIVATE_INDEX_ATTRUSED))))

	metas, ok := Restore(fdwPrivate).([]*TargetColumnMeta)
	if (!ok) {
		/*
		   the long-jump in ereport will directly jump into the Postgres's C-stack, 
		   I'm not sure if the machanism of "defer" in Go would take effect 
		 */		
		Unref(fdwPrivate)
		ereport(ERRCODE_FDW_ERROR, "internal error: type assersion of \"%p\" to \"[]*TargetColumnMeta\" failed ", fdwPrivate)
	}
	Unref(fdwPrivate)

	dbstate := new(DoubanScanState)
	dbstate.scanningRel = rel
	dbstate.attrsRetrieved = metas
	dbstate.resultSet = items
	dbstate.currentRow = 0 // zero means "having not started iterating yet"
	node.fdw_state = Save(dbstate)
}

//export doubanIterateForeignScan
func doubanIterateForeignScan(node *C.ForeignScanState) *C.TupleTableSlot {
	iterateMax := MovieRankingTop250Num
	sstate := (*C.ScanState)(unsafe.Pointer(&(node.ss)))
	slot := (*C.TupleTableSlot)(unsafe.Pointer(sstate.ss_ScanTupleSlot))
	tupDesc := (C.TupleDesc)(unsafe.Pointer(slot.tts_tupleDescriptor))
	dbstate, ok := Restore(unsafe.Pointer(node.fdw_state)).(*DoubanScanState)
	if (!ok) {
		/*
		   the long-jump in ereport will directly jump into the Postgres's C-stack, 
		   I'm not sure if the machanism of "defer" in Go would take effect 
		 */
		ereport(ERRCODE_FDW_ERROR, "internal error: type assersion of \"%p\" to \"*DoubanScanState\" failed ", 
			unsafe.Pointer(node.fdw_state))
	}

	if len(dbstate.resultSet) < iterateMax {
		iterateMax = len(dbstate.resultSet)
	}

	if dbstate.currentRow < 0 || dbstate.currentRow > iterateMax {
		ereport(ERRCODE_FDW_ERROR, "internal error: \"DoubanScanState.currentRow\" %d beyond the upper value %d", 
			dbstate.currentRow, iterateMax)
	}

	natts := int(tupDesc.natts)

	C.memset(unsafe.Pointer(slot.tts_values), 0x00, C.size_t(C.sizeof_Datum * tupDesc.natts))
	C.memset(unsafe.Pointer(slot.tts_isnull), C.int(1), C.size_t(C.sizeof_bool * tupDesc.natts))
	C.ExecClearTuple(slot)
	/*
	 * all the Top250 items were retrieved, it's time to stop the iteration
	 * TODO: if the limit clause pushdown were implemented, the following if-condition
	*        should be modified  
	 */
	if dbstate.currentRow == iterateMax {
		return slot
	}

	datumsSlice := (*[1 << 30]C.Datum)(unsafe.Pointer(slot.tts_values))[:natts:natts]
	isnullsSlice := (*[1 << 30]C.bool)(unsafe.Pointer(slot.tts_isnull))[:natts:natts]
	for _, val := range dbstate.attrsRetrieved {
		isnullsSlice[val.attrNum - 1] = C.bool(0)
		datumsSlice[val.attrNum - 1] = val.convert2Datum(&(dbstate.resultSet[dbstate.currentRow]))
	}
	C.ExecStoreVirtualTuple(slot);
	dbstate.currentRow += 1

	return slot
}

//export doubanEndForeignScan
func doubanEndForeignScan(node *C.ForeignScanState) {
	Unref(unsafe.Pointer(node.fdw_state))
	node.fdw_state = nil
}

//export doubanReScanForeignScan
func doubanReScanForeignScan(node *C.ForeignScanState) {
	dbstate, ok := Restore(unsafe.Pointer(node.fdw_state)).(*DoubanScanState)
	if (!ok) {
		/*
		   the long-jump in ereport will directly jump into the Postgres's C-stack, 
		   I'm not sure if the machanism of "defer" in Go would take effect 
		 */
		ereport(ERRCODE_FDW_ERROR, "internal error: type assersion of \"%p\" to \"*DoubanScanState\" failed ", 
			unsafe.Pointer(node.fdw_state))
	}

	// reset the current row count
	if dbstate.resultSet != nil {
		dbstate.currentRow = 0
	} else {
		sstate := (*C.ScanState)(unsafe.Pointer(&(node.ss)))
		rel := (*C.RelationData)(unsafe.Pointer(sstate.ss_currentRelation))
	
		rank := getRankNameFromForeginTable(rel)

		items, err := RetrieveRankingData(rank, 50) //TODO: constant variable 50 should be changed
		if err != nil {
			ereport(ERRCODE_FDW_ERROR,
				"error occurred while retrieving data from douban.com\n  details:%v", err)
			return // ereport will cause the statement jumped out of the execution.
		}
		dbstate.resultSet = items
		dbstate.currentRow = 0
	}
}

//export doubanExplainForeignScan
func doubanExplainForeignScan(node *C.ForeignScanState,
	es *C.ExplainState) {

	title := C.CString("Douban Rank")
	defer C.free(unsafe.Pointer(title))

	sstate := (*C.ScanState)(unsafe.Pointer(&(node.ss)))
	rel := (*C.RelationData)(unsafe.Pointer(sstate.ss_currentRelation))
	rankName := getRankNameFromForeginTable(rel)

	C.ExplainPropertyText(title, C.CString(rankName), es)

	if int(es.costs) > 0 {
		detailTitle := C.CString("Movie items")

		C.ExplainPropertyLong(detailTitle, C.long(MovieRankingTop250Num), es)
	}
}

//export doubanAnalyzeForeignTable
func doubanAnalyzeForeignTable(relation *C.RelationData,
	aquireSampleRowsFunc *C.AcquireSampleRowsFunc, totalpages *C.uint) C.bool {
	*totalpages = 1
	return C.bool(0)    /* the foreign data cannot be sampled since it's a web api */
}

//export checkOptionName
func checkOptionName(optname *C.char, context C.Oid) C.int {
	opt := strings.ToLower(C.GoString(optname))

	if v, ok := ValidOptions[opt]; ok {
		if v.oid == context {
			return v.identifier
		}
	}
	return C.int(-1)
}

//export checkRankName
func checkRankName(rankname *C.char) C.bool {
	rank := strings.ToLower(C.GoString(rankname))

	if _, ok := UrlMap[rank]; ok {
		return C.bool(1)
	}
	return C.bool(0)
}

func getRankNameFromForeginTable(rel *C.RelationData) string {
	var reloid C.Oid
	var table *C.ForeignTable

	reloid = rel.rd_id
	table = C.GetForeignTable(reloid)
	if table == nil {
		return ""
	}

	optionCount := int(C.list_length(table.options))
    for i := 0; i < optionCount; i++ {
		def := (*C.DefElem)(unsafe.Pointer(uintptr(C.list_nth(table.options, C.int(i)))))
		if strings.ToLower(C.GoString(def.defname)) == "rank_name" {
			// we don't need to worry about the correctness of def value
			// because a validation has been executed during CREATE FOREIGN TABLE
			return C.GoString(C.defGetString(def))
		}
	}

	/* "rank_name" option not found */
	relationName := (* C.NameData)(unsafe.Pointer(&(((*C.FormData_pg_class)(unsafe.Pointer(rel.rd_rel))).relname)))
	tabname := C.GoString(&(relationName.data[0]))
	ereport(ERRCODE_FDW_INVALID_OPTION_NAME, 
		"the \"rank_name\" option not specified while defining the foreign table \"%s\"", tabname)

	return ""    /* avoid the compile error */
}

func main() {}
