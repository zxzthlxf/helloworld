package pq

import (
	"reflect"
	"unsafe"
)

func GetCString(str *string) (str_cstring *Cchar, str_clen Cint) {

	str_header := (*reflect.StringHeader)(unsafe.Pointer(str))
	str_cstring = (*Cchar)(unsafe.Pointer(str_header.Data))
	str_clen = Cint(str_header.Len)
	
	return str_cstring, str_clen
}
