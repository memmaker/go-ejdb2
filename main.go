package ejdb2

// #cgo LDFLAGS: -lejdb2
// #include <stdlib.h>
// #include <ejdb2/ejdb2.h>
// #include <ejdb2/iowow/iwkv.h>
// extern void goVisitor(EJDB_DOC doc, long long *step);
// static inline unsigned long long goVisitorWrapper(EJDB_EXEC *ctx, EJDB_DOC doc, long long *step) {
//   goVisitor(doc, step);
//   return 0;
// }
// static iwrc execute_query(EJDB db, JQL q) {
//   EJDB_EXEC ux = {
//     .db = db,
//     .q = q,
//     .visitor = goVisitorWrapper
//   };
//   return ejdb_exec(&ux);
// }
// static IWXSTR* jbl_to_json(JBL jbl) {
//   IWXSTR *xstr = iwxstr_new();
//   iwrc rc = jbl_as_json(jbl, jbl_xstr_json_printer, xstr, JBL_PRINT_PRETTY);
//   return xstr;
// }
import "C"
import (
	"encoding/json"
	"fmt"
	"unsafe"
)

type J = map[string]interface{}

type EJDB struct {
	db C.EJDB
}

var visitorCallback func(jsonRecord string)

//export goVisitor
func goVisitor(doc C.EJDB_DOC, step *C.longlong) {
	output := jblTojson(doc.raw)
	if visitorCallback != nil {
		visitorCallback(output)
	}
}

func Check(errorCode C.iwrc) error {
	if errorCode != 0 {
		return fmt.Errorf("error code: %d", int(errorCode))
	}
	return nil
}

func (e *EJDB) Open(filename string) error {
	filenameCString := C.CString(filename)
	opts := C.EJDB_OPTS{
		kv: C.IWKV_OPTS{
			path: filenameCString,
			//oflags: C.IWKV_TRUNC,
		},
	}
	defer func() {
		C.free(unsafe.Pointer(filenameCString))
		//C.free(unsafe.Pointer(&opts))
	}()

	rc := C.ejdb_init()
	err := Check(rc)
	if err != nil {
		return err
	}
	rc = C.ejdb_open(&opts, &e.db)
	return Check(rc)
}

func (e *EJDB) Close() {
	C.ejdb_close(&e.db)
}

func (e *EJDB) GetMeta() string {
	var meta C.JBL
	rc := C.ejdb_get_meta(e.db, &meta)
	defer C.jbl_destroy(&meta)
	err := Check(rc)
	if err != nil {
		print(err)
		return ""
	}
	jsonString := jblTojson(meta)
	return jsonString
}

func (e *EJDB) GetCollections() []string {
	var collectionNames []string
	meta := J{}
	metaString := e.GetMeta()
	if metaString == "" {
		return collectionNames
	}
	err := json.Unmarshal([]byte(metaString), &meta)
	if err != nil {
		return collectionNames
	}
	for _, collectionJson := range meta["collections"].([]interface{}) {
		coll := collectionJson.(J)
		collectionNames = append(collectionNames, coll["name"].(string))
	}
	return collectionNames
}

func (e *EJDB) MergeOrPut(collectionName string, patchJSON string, entryID int64) error {
	patchJSONCString := C.CString(patchJSON)
	collectionCString := C.CString(collectionName)

	defer func() {
		C.free(unsafe.Pointer(patchJSONCString))
		C.free(unsafe.Pointer(collectionCString))
	}()

	rc := C.ejdb_merge_or_put(e.db, collectionCString, patchJSONCString, C.int64_t(entryID))
	return Check(rc)
}

func (e *EJDB) PutNew(collectionName string, jsonData string) (int64, error) {
	var jbl C.JBL
	var id C.int64_t
	jsonDataCString := C.CString(jsonData)
	collectionCString := C.CString(collectionName)

	defer func() {
		C.jbl_destroy(&jbl)
		C.free(unsafe.Pointer(jsonDataCString))
		C.free(unsafe.Pointer(collectionCString))
	}()

	rc := C.jbl_from_json(&jbl, jsonDataCString)
	err := Check(rc)
	if err != nil {
		return -1, err
	}
	rc = C.ejdb_put_new(e.db, collectionCString, jbl, &id)

	return int64(id), Check(rc)
}

func (e *EJDB) Put(collectionName string, jsonData string, id int64) error {
	var jbl C.JBL
	jsonDataCString := C.CString(jsonData)
	collectionCString := C.CString(collectionName)

	defer func() {
		C.jbl_destroy(&jbl)
		C.free(unsafe.Pointer(jsonDataCString))
		C.free(unsafe.Pointer(collectionCString))
	}()

	rc := C.jbl_from_json(&jbl, jsonDataCString)
	err := Check(rc)
	if err != nil {
		return err
	}
	rc = C.ejdb_put(e.db, collectionCString, jbl, C.int64_t(id))
	err = Check(rc)
	return err
}

func (e *EJDB) Patch(collectionName string, patchJSON string, entryID int64) error {
	collectionCString := C.CString(collectionName)
	pathJSONCString := C.CString(patchJSON)
	defer func() {
		C.free(unsafe.Pointer(collectionCString))
		C.free(unsafe.Pointer(pathJSONCString))
	}()
	rc := C.ejdb_patch(e.db, collectionCString, pathJSONCString, C.int64_t(entryID))
	return Check(rc)
}

func (e *EJDB) Update(collectionName string, query string, params J) error {
	q, free := newQuery(collectionName, query, params)
	defer free()
	rc := C.ejdb_update(e.db, q)
	return Check(rc)
}

func (e *EJDB) Del(collectionName string, entryID int64) error {
	collectionCString := C.CString(collectionName)
	defer C.free(unsafe.Pointer(collectionCString))
	rc := C.ejdb_del(e.db, collectionCString, C.int64_t(entryID))
	return Check(rc)
}

func (e *EJDB) GetByID(collectionName string, id int64) string {
	var jbl C.JBL
	collectionCString := C.CString(collectionName)
	defer func() {
		C.jbl_destroy(&jbl)
		C.free(unsafe.Pointer(collectionCString))
	}()

	rc := C.ejdb_get(e.db, collectionCString, C.int64_t(id), &jbl)
	err := Check(rc)
	if err != nil {
		return ""
	}
	jsonString := jblTojson(jbl)
	return jsonString
}

func (e *EJDB) Get(collectionName string, query string, params J, visitor func(string)) error {
	var resultList C.EJDB_LIST
	visitorCallback = visitor
	q, free := newQuery(collectionName, query, params)
	queryCString := C.CString(query)
	collectionCString := C.CString(collectionName)
	defer func() {
		free()
		C.ejdb_list_destroy(&resultList)
		C.free(unsafe.Pointer(collectionCString))
		C.free(unsafe.Pointer(queryCString))
	}()

	rc := C.execute_query(e.db, q)
	//rc := C.ejdb_list2(e.db, collectionCString, queryCString, limit, &resultList)
	return Check(rc)
}

func (e *EJDB) Count(collectionName string, query string, params J) int64 {
	q, free := newQuery(collectionName, query, params)
	defer free()
	var countValue C.longlong
	rc := C.ejdb_count(e.db, q, &countValue, 0)
	Check(rc)
	return int64(countValue)
}

func (e *EJDB) EnsureCollection(collectionName string) error {
	collectionCString := C.CString(collectionName)
	defer C.free(unsafe.Pointer(collectionCString))
	rc := C.ejdb_ensure_collection(e.db, collectionCString)
	return Check(rc)
}

func (e *EJDB) RemoveCollection(collectionName string) error {
	collectionCString := C.CString(collectionName)
	defer C.free(unsafe.Pointer(collectionCString))
	rc := C.ejdb_remove_collection(e.db, collectionCString)
	return Check(rc)
}

func (e *EJDB) RenameCollection(oldName string, newName string) error {
	oldNameCString := C.CString(oldName)
	newNameCString := C.CString(newName)
	defer func() {
		C.free(unsafe.Pointer(oldNameCString))
		C.free(unsafe.Pointer(newNameCString))
	}()
	rc := C.ejdb_rename_collection(e.db, oldNameCString, newNameCString)
	return Check(rc)
}

func (e *EJDB) EnsureIndex(collectionName string, path string, mode IndexMode) error {
	collectionCString := C.CString(collectionName)
	pathCString := C.CString(path)
	defer func() {
		C.free(unsafe.Pointer(collectionCString))
		C.free(unsafe.Pointer(pathCString))
	}()
	rc := C.ejdb_ensure_index(e.db, collectionCString, pathCString, C.ejdb_idx_mode_t(mode))
	return Check(rc)
}

func (e *EJDB) RemoveIndex(collectionName string, path string, mode IndexMode) error {
	collectionCString := C.CString(collectionName)
	pathCString := C.CString(path)
	defer func() {
		C.free(unsafe.Pointer(collectionCString))
		C.free(unsafe.Pointer(pathCString))
	}()
	rc := C.ejdb_remove_index(e.db, collectionCString, pathCString, C.ejdb_idx_mode_t(mode))
	return Check(rc)
}

func (e *EJDB) OnlineBackup(targetFile string) (uint64, error) {
	var timeStamp C.ulonglong
	pathCString := C.CString(targetFile)
	defer func() {
		C.free(unsafe.Pointer(pathCString))
	}()
	rc := C.ejdb_online_backup(e.db, &timeStamp, pathCString)
	return uint64(timeStamp), Check(rc)
}

type IndexMode uint8

const (
	Unique  uint8 = 0x01
	String        = 0x04
	Integer       = 0x08
	Float         = 0x10
)

func newQuery(collectionName string, query string, params J) (C.JQL, func()) {
	var q C.JQL
	var allocatedPointers []unsafe.Pointer
	freeFunc := func() {
		for _, pointer := range allocatedPointers {
			C.free(pointer)
		}
		C.jql_destroy(&q)
	}
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("panic occurred during query building:", err)
			fmt.Println("freeing allocated pointers")
			freeFunc()
		}
	}()

	collectionCString := C.CString(collectionName)
	queryCString := C.CString(query)

	allocatedPointers = append(allocatedPointers, unsafe.Pointer(collectionCString))
	allocatedPointers = append(allocatedPointers, unsafe.Pointer(queryCString))

	rc := C.jql_create(&q, collectionCString, queryCString)
	Check(rc)

	for key, value := range params {
		// NOTE: could probably be optimized by just passing the type along
		keyCString := C.CString(key)
		allocatedPointers = append(allocatedPointers, unsafe.Pointer(keyCString))

		switch v := value.(type) {
		case int:
			rc = C.jql_set_i64(q, keyCString, 0, C.longlong(v))
			break
		case int64:
			rc = C.jql_set_i64(q, keyCString, 0, C.longlong(v))
			break
		case float32:
			rc = C.jql_set_f64(q, keyCString, 0, C.double(v))
			break
		case float64:
			rc = C.jql_set_f64(q, keyCString, 0, C.double(v))
			break
		case string:
			allocatedString := C.CString(v)
			allocatedPointers = append(allocatedPointers, unsafe.Pointer(allocatedString))
			rc = C.jql_set_str(q, keyCString, 0, allocatedString)
			break
		case bool:
			rc = C.jql_set_bool(q, keyCString, 0, value.(C.bool))
			break
		}
		Check(rc)
	}

	return q, freeFunc
}

func jblTojson(jbl C.JBL) string {
	xstr := C.jbl_to_json(jbl)
	jsonString := C.iwxstr_ptr(xstr)
	defer C.iwxstr_destroy(xstr)
	jsonGO := C.GoString(jsonString)
	return jsonGO
}
