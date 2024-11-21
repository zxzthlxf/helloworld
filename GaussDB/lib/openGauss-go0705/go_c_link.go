//go:build enable_ce
// +build enable_ce

package pq

/*
#cgo LDFLAGS: -ldl
#include <stdlib.h>
#include "clientlogic_hooks.h"
#include "pgconn.h"
#include "lib_funcs_manager.h"
typedef struct pg_conn PGconn;
*/
import "C"
import "unsafe"

type Cchar = C.char
type Cint = C.int
type Csize_t = C.size_t
type PGconn = C.PGconn
type StatementData = C.StatementData
type Cuint = C.uint
type Culong = C.ulong
type CLRefreshParams = C.CLRefreshParams
type PGClientLogic = C.PGClientLogic
type Coid = C.uint

func bool_to_Cint(b bool) Cint {
	if b {
		return Cint(1)
	} else {
		return Cint(0)
	}
}

func Cint_to_bool(i Cint) bool {
	return i != Cint(0)
}

func Cmalloc(sz Csize_t) unsafe.Pointer {
	return C.malloc(sz)
}

func CString(str string) *Cchar {
	return C.CString(str)
}

func GoString(str *Cchar) string {
	return C.GoString(str)
}

func Cfree(ptr unsafe.Pointer) {
	C.free(ptr)
}

func variadic_Cfree(ptrs ...unsafe.Pointer) {
    for _, ptr := range ptrs {
        C.free(ptr)
    }
}

func GoBytes(ptr unsafe.Pointer, len Cint) []byte {
	return C.GoBytes(ptr, len)
}

func CBytes(bytes []byte) unsafe.Pointer {
	return C.CBytes(bytes)
}

func run_post_query(conn *PGconn, rfq_cache_refresh_type Cint, conn_status bool) Cint {
	return C.run_post_query(conn, rfq_cache_refresh_type, bool_to_Cint(conn_status))
}

func run_pre_query(conn *PGconn, stmt_name *Cchar, query *Cchar, client_side_error *Cint) *StatementData {
	return C.run_pre_query(conn, stmt_name, query, client_side_error)
}

func run_pre_exec(conn *PGconn, stmt_name *Cchar, params_count Csize_t, params_values **Cchar,
	params_lengths *Cint, params_formats *Cint, client_side_error *Cint) *StatementData {
	return C.run_pre_exec(conn, stmt_name, params_count, params_values, params_lengths, params_formats, client_side_error)
}

func accept_pending_statements(conn *PGconn) Cint {
	return C.accept_pending_statements(conn)
}

func free_mem_manager() {
	C.free_mem_manager()
}

func deprocess_value(processed_value *Cchar, processed_value_len Cint, original_type Cint,
	length_output *Csize_t, output_binary_format Cint, pgconn *PGconn) *Cchar {
	return C.deprocess_value(processed_value, processed_value_len, original_type,
		length_output, output_binary_format, pgconn)
}

func deprocess_record(processed_value *Cchar, processed_value_len Cint, original_type Cint, length_output *Csize_t,
	function_name *Cchar, output_binary_format Cint, ttype Cuint, pgconn *PGconn) *Cchar {
	return C.deprocess_record(processed_value, processed_value_len, original_type, length_output,
		function_name, output_binary_format, ttype, pgconn)
}

func statement_data_get_query(statement_data *StatementData) *Cchar {
	return C.statement_data_get_query(statement_data)
}

func pgconn_errmsg(pgconn *PGconn) *Cchar {
	return C.pgconn_errmsg(pgconn)
}

func delete_statementdata_c(statement_data *StatementData) {
	C.delete_statementdata_c(statement_data)
}

func post_check_resend_query_on_error(conn *PGconn) {
	C.post_check_resend_query_on_error(conn)
}

func pre_check_resend_query_on_error(conn *PGconn) bool {
	return Cint_to_bool(C.pre_check_resend_query_on_error(conn))
}

func is_any_refresh_cache_on_error(conn *PGconn) bool {
	return Cint_to_bool(C.is_any_refresh_cache_on_error(conn))
}

func pgconn_init(go_conn unsafe.Pointer, user *Cchar, user_len Cint,
	database *Cchar, database_len Cint,
	application_name *Cchar, application_name_len Cint, enable_full_encryption Cint) *PGconn {
	return C.pgconn_init(go_conn, user, user_len, database, database_len, application_name, application_name_len, enable_full_encryption)
}

func pgconn_free(conn *PGconn) {
	C.pgconn_free(conn)
}

func pgconn_setserverversion(conn *PGconn, server_version Cint) {
	C.pgconn_setserverversion(conn, server_version)
}

func pgconn_setstdstrings(conn *PGconn, std_strings Cint) {
	C.pgconn_setstdstrings(conn, std_strings)
}

func pgconn_reset(conn *PGconn) {
	C.pgconn_reset(conn)
}

func check_library() bool {
	return Cint_to_bool(C.check_library())
}

func deprocess_error_detail_c(conn *PGconn, value *Cchar, contents *Cchar) bool {
	return Cint_to_bool(C.deprocess_error_detail_c(conn, value, contents))
}

func get_last_refresh_cache_status(conn *PGconn, r_error **Cchar) bool {
	return Cint_to_bool(C.get_last_refresh_cache_status(conn, r_error))
}

func statement_data_get_params_count(statement_data *StatementData) Csize_t {
	return C.statement_data_get_params_count(statement_data)
}

func processed_query_pos_to_deprocessed_c(conn *PGconn, processed_pos Cint) Cint {
	return C.processed_query_pos_to_deprocessed_c(conn, processed_pos)
}

func statement_data_get_params_values(statement_data *StatementData) **Cchar {
	return C.statement_data_get_params_values(statement_data)
}

func statement_data_get_params_lengths(statement_data *StatementData) *Cint {
	return C.statement_data_get_params_lengths(statement_data)
}

func process_copy_chunk(conn *PGconn, in_buffer *Cchar, msg_length Cint, buffer **Cchar) Cint {
	return C.process_copy_chunk(conn, in_buffer, msg_length, buffer)
}

func new_cl_refresh_params() *CLRefreshParams {
	return C.new_cl_refresh_params()
}

func get_client_logic(conn *PGconn) *PGClientLogic {
	return C.get_client_logic(conn)
}

func clientlogic_read_error(client_logic *PGClientLogic, id Cchar, data *Cchar, cl_refresh_params *CLRefreshParams) {
	C.clientlogic_read_error(client_logic, id, data, cl_refresh_params)
}

func delete_cl_refresh_params(cl_refresh_params *CLRefreshParams) {
	C.delete_cl_refresh_params(cl_refresh_params)
}

//export Conncgo_SimpleQuery
func Conncgo_SimpleQuery(conn_ptr unsafe.Pointer, query string) (result ***Cchar,
	columns_names **Cchar, rows_count int, columns_count int) {
	return Conncgo_SimpleQuery_func(conn_ptr, query)
}

func Is_built_with_cgo() bool {
	return true
}

func clear_client_keys(conn *PGconn) bool {
	return Cint_to_bool(C.clear_client_keys_c(conn))
}

func set_key_to_client_logic(conn *PGconn, rsa_key_len Csize_t, ecdh_key_len Csize_t, rsa_pub_key *Cchar, ecdh_key *Cchar) bool {
	return Cint_to_bool(C.set_key_to_client_logic_c(conn, rsa_key_len, ecdh_key_len, rsa_pub_key, ecdh_key))
}

func get_client_key(conn *PGconn, client_key **Cchar, client_key_len *Csize_t) bool {
	return Cint_to_bool(C.get_client_key_c(conn, client_key, client_key_len))
}

func fetched_columns(conn *PGconn) Cint {
	return C.fetched_columns_c(conn)
}

func get_cached_ceks(conn *PGconn, cek_oids **Coid, cek_algos **Cint, cek_buf_lens **Cint, cek_bufs ***Cchar) bool {
	return Cint_to_bool(C.get_cached_ceks_c(conn, cek_oids, cek_algos, cek_buf_lens, cek_bufs))
}
