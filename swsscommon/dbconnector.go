package swsscommon

// #cgo LDFLAGS: -lcswsscommon -lswsscommon -lstdc++
// #include <capi/dbconnector.h>
// #include <swss/c-api/dbconnector.h>
// #include <stdlib.h>
import "C"

import (
    "unsafe"
    "errors"
)

type DBConnector struct {
    ptr unsafe.Pointer
}

func NewDBConnector(db int, hostname string, port int, timeout uint) DBConnector {
    hostnameC := C.CString(hostname)
    defer C.free(unsafe.Pointer(hostnameC))
    dbc := C.db_connector_new(C.int(db), hostnameC, C.int(port), C.uint(timeout))
    return DBConnector{ptr: unsafe.Pointer(dbc)}
}

func NewDBConnector2(db int, unixPath string, timeout uint) DBConnector {
    unixPathC := C.CString(unixPath)
    defer C.free(unsafe.Pointer(unixPathC))
    dbc := C.db_connector_new2(C.int(db), unixPathC, C.uint(timeout))
    return DBConnector{ptr: unsafe.Pointer(dbc)}
}

func NewDBConnectorNamed(name string, timeout uint, authenticate bool) (DBConnector, error) {
    nameC := C.CString(name)
    defer C.free(unsafe.Pointer(nameC))
    var enabled C.uint8_t
    if authenticate { enabled = 1 }
    var dbc C.SWSSDBConnector
    result := C.SWSSDBConnector_new_named_with_auth(nameC, C.uint32_t(timeout), 0, enabled, &dbc)
    defer C.SWSSString_free(result.message)
    defer C.SWSSString_free(result.location)
    if result.exception != C.SWSSException_None {
        return DBConnector{}, errors.New(C.GoString(C.SWSSStrRef_c_str(C.SWSSStrRef(unsafe.Pointer(result.message)))))
    }
    return DBConnector{ptr: unsafe.Pointer(dbc)}, nil
}

func (db DBConnector) Delete() {
    C.db_connector_delete(C.db_connector_t(db.ptr))
}

func (db DBConnector) GetDB() int {
    return int(C.db_connector_get_db(C.db_connector_t(db.ptr)))
}

func DBConnectorSelect(db DBConnector) {
    C.db_connector_select(C.db_connector_t(db.ptr))
}

func (db DBConnector) NewConnector(timeout uint) DBConnector {
    dbc := C.db_connector_new_connector(C.db_connector_t(db.ptr), C.uint(timeout))
    return DBConnector{ptr: unsafe.Pointer(dbc)};
}
