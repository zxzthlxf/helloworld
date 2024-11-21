//go:build !enable_ce
// +build !enable_ce

package pq

import "unsafe"

type Cchar = string
type Cint = int
type Csize_t = uint32
type PGconn = unsafe.Pointer
type StatementData = unsafe.Pointer
type Cuint = uint
type Culong = uint32
type CLRefreshParams = unsafe.Pointer
type PGClientLogic = unsafe.Pointer
type Coid = uint

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
	panic("Bad function call Cmalloc")
}

func CString(str string) *Cchar {
	panic("Bad function call CString")
}

func GoString(str *Cchar) string {
	panic("Bad function call GoString")
}

func Cfree(ptr unsafe.Pointer) {
	panic("Bad function call Cfree")
}

func variadic_Cfree(ptrs ...unsafe.Pointer) {
	panic("Bad function call variadic_Cfree")
}

func GoBytes(ptr unsafe.Pointer, len Cint) []byte {
	panic("Bad function call GoBytes")
}

func CBytes(bytes []byte) unsafe.Pointer {
	panic("Bad function call CBytes")
}

func run_post_query(conn *PGconn, rfq_cache_refresh_type Cint, conn_status bool) Cint {
	panic("Bad function call run_post_query")
}

func run_pre_query(conn *PGconn, stmt_name *Cchar, query *Cchar, client_side_error *Cint) *StatementData {
	panic("Bad function call run_pre_query")
}

func run_pre_exec(conn *PGconn, stmt_name *Cchar, params_count Csize_t, params_values **Cchar,
	params_lengths *Cint, params_formats *Cint, client_side_error *Cint) *StatementData {
	panic("Bad function call run_pre_exec")
}

func accept_pending_statements(conn *PGconn) Cint {
	panic("Bad function call accept_pending_statements")
}

func free_mem_manager() {
	panic("Bad function call free_mem_manager")
}

func deprocess_value(processed_value *Cchar, processed_value_len Cint, original_type Cint,
	length_output *Csize_t, output_binary_format Cint, pgconn *PGconn) *Cchar {
	panic("Bad function call deprocess_value")
}

func deprocess_record(processed_value *Cchar, processed_value_len Cint, original_type Cint, length_output *Csize_t,
	function_name *Cchar, output_binary_format Cint, ttype Cuint, pgconn *PGconn) *Cchar {
	panic("Bad function call deprocess_record")
}

func statement_data_get_query(statement_data *StatementData) *Cchar {
	panic("Bad function call statement_data_get_query")
}

func pgconn_errmsg(pgconn *PGconn) *Cchar {
	panic("Bad function call pgconn_errmsg")
}

func delete_statementdata_c(statement_data *StatementData) {
	panic("Bad function call delete_statementdata_c")
}

func post_check_resend_query_on_error(conn *PGconn) {
	panic("Bad function call post_check_resend_query_on_error")
}

func pre_check_resend_query_on_error(conn *PGconn) bool {
	panic("Bad function call pre_check_resend_query_on_error")
}

func is_any_refresh_cache_on_error(conn *PGconn) bool {
	panic("Bad function call is_any_refresh_cache_on_error")
}

func pgconn_init(go_conn unsafe.Pointer, user *Cchar, user_len Cint,
	database *Cchar, database_len Cint,
	application_name *Cchar, application_name_len Cint, enable_full_encryption Cint) *PGconn {
	panic("Bad function call pgconn_init")
}

func pgconn_free(conn *PGconn) {
	panic("Bad function call pgconn_free")
}

func pgconn_setserverversion(conn *PGconn, server_version Cint) {
	panic("Bad function call pgconn_setserverversion")
}

func pgconn_setstdstrings(conn *PGconn, std_strings Cint) {
	panic("Bad function call pgconn_setstdstrings")
}

func pgconn_reset(conn *PGconn) {
	panic("Bad function call pgconn_reset")
}

func check_library() bool {
	return false
}

func deprocess_error_detail_c(conn *PGconn, value *Cchar, contents *Cchar) bool {
	panic("Bad function call deprocess_error_detail_c")
}

func get_last_refresh_cache_status(conn *PGconn, r_error **Cchar) bool {
	panic("Bad function call get_last_refresh_cache_status")
}

func statement_data_get_params_count(statement_data *StatementData) Csize_t {
	panic("Bad function call statement_data_get_params_count")
}

func processed_query_pos_to_deprocessed_c(conn *PGconn, processed_pos Cint) Cint {
	panic("Bad function call processed_query_pos_to_deprocessed_c")
}

func statement_data_get_params_values(statement_data *StatementData) **Cchar {
	panic("Bad function call statement_data_get_params_values")
}

func statement_data_get_params_lengths(statement_data *StatementData) *Cint {
	panic("Bad function call statement_data_get_params_lengths")
}

func process_copy_chunk(conn *PGconn, in_buffer *Cchar, msg_length Cint, buffer **Cchar) Cint {
	panic("Bad function call process_copy_chunk")
}

func new_cl_refresh_params() *CLRefreshParams {
	panic("Bad function call new_cl_refresh_params")
}

func get_client_logic(conn *PGconn) *PGClientLogic {
	panic("Bad function call get_client_logic")
}

func clientlogic_read_error(client_logic *PGClientLogic, id Cchar, data *Cchar, cl_refresh_params *CLRefreshParams) {
	panic("Bad function call clientlogic_read_error")
}

func delete_cl_refresh_params(cl_refresh_params *CLRefreshParams) {
	panic("Bad function call delete_cl_refresh_params")
}

//export Conncgo_SimpleQuery
func Conncgo_SimpleQuery(conn_ptr unsafe.Pointer, query string) (result ***Cchar,
	columns_names **Cchar, rows_count int, columns_count int) {
	panic("Bad function call Conncgo_SimpleQuery")
}

func Is_built_with_cgo() bool {
	return false
}

func clear_client_keys(conn *PGconn) bool {
	panic("Bad function call clear_client_keys")
}

func set_key_to_client_logic(conn *PGconn, rsa_key_len Csize_t, ecdh_key_len Csize_t, rsa_pub_key *Cchar,
	ecdh_key *Cchar) bool {
	panic("Bad function call set_key_to_client_logic")
}

func get_client_key(conn *PGconn, client_key **Cchar, client_key_len *Csize_t) bool {
	panic("Bad function call get_client_key")
}

func fetched_columns(conn *PGconn) Cint {
	panic("Bad function call fetched_columns")
}

func get_cached_ceks(conn *PGconn, cek_oids **Coid, cek_algos **Cint, cek_buf_lens **Cint, cek_bufs ***Cchar) bool {
	panic("Bad function call get_cache_ceks")
}

func set_key_info(conn *PGconn, key_info *Cchar) bool {
	panic("Bad function call set_key_info")
}

func set_crypto_module_info(conn *PGconn, crypto_module_info *Cchar) bool {
	panic("Bad function call set_crypto_module_info")
}
