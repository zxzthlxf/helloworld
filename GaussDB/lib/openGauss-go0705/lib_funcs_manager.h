/*
 * Copyright (c) 2023 Huawei Technologies Co.,Ltd.
 *
 * openGauss is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *
 * http://license.coscl.org.cn/MulanPSL2
 *
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 * -------------------------------------------------------------------------
 *
 * lib_funcs_manager.h
 *
 * IDENTIFICATION
 * 	  src/gitee.com/opengauss/openGauss-connector-go-pq/lib_funcs_manager.h
 *
 * -------------------------------------------------------------------------
 */
#ifndef LIB_FUNCS_MANAGER_H
#define LIB_FUNCS_MANAGER_H

#include <string.h>
typedef unsigned int Oid;

#ifdef __cplusplus
extern "C"
{
#endif // __cplusplus
    typedef struct PQExpBufferData PQExpBufferData;
    typedef PQExpBufferData *PQExpBuffer;
    typedef struct pg_conn PGconn;
    typedef struct StatementData StatementData;
    typedef struct PGClientLogic PGClientLogic;
    typedef struct CursorInterface CursorInterface;
    typedef struct Deprocess Deprocess;
    typedef struct CLRefreshParams CLRefreshParams;
    typedef void (*CursorDeallocator)(CursorInterface *);

    int run_post_query_c(PGconn *, int);
    const int *get_rec_origial_ids(PGconn *, unsigned int, char const *);
    size_t get_rec_origial_ids_length(PGconn *, unsigned int, char const *);
    int de_process_record(PGconn *, char const *, unsigned long, int const *, unsigned long, int, unsigned char **, unsigned long *, int *);
    Deprocess *new_deprocess(int, int, int);
    void delete_deprocess(Deprocess *);
    void deprocess_set_conn_c(Deprocess *deprocess, PGconn *);
    int get_deprocessed_c(Deprocess *deprocess, const unsigned char *processed_data, size_t processed_data_size,
                          unsigned int type_oid, unsigned char **plain_text, size_t *plain_text_size);
    int processor_run_pre_query_c(StatementData *, int is_inner_query, int *failed_to_parse);
    int run_pre_exec_c(StatementData *);
    void delete_statementdata_c(StatementData *);
    int accept_pending_statements_c(PGconn *, int);
    void free_mem_manager();
    int process_copy_chunk(PGconn *conn, const char *in_buffer, int msg_length, char **buffer);
    int get_enum_val_by_string(const char *str);
    void post_check_resend_query_on_error(PGconn *conn);
    int pre_check_resend_query_on_error(PGconn *conn);
    int is_any_refresh_cache_on_error(PGconn *conn);
    void clientlogic_read_error(PGClientLogic *client_logic, const char id, const char *data,
        CLRefreshParams *cl_refresh_params);
    int deprocess_error_detail_c(PGconn *conn, const char *value, char *contents);
    StatementData *new_statementdata_c(PGconn *aconn, const char *astmt_name, const char *aquery,
        const size_t an_params, const Oid *aparam_types, const char * const * aparam_values, const int *aparam_lengths,
        const int *aparam_formats, int is_direct_stmt_flow);
    size_t statement_data_get_params_count(StatementData *statement_data);
    int processed_query_pos_to_deprocessed_c(PGconn *conn, int processed_pos);
    const char * const * statement_data_get_params_values(StatementData *statement_data);
    int *statement_data_get_params_lengths(StatementData *statement_data);
    const char *statement_data_get_query(StatementData *statement_data);
    void set_cl_rfq_cache_refresh_type(PGconn *conn, int rfq);
    void set_conn_status(PGconn *conn, int status);
    int is_process_query(PGconn *conn);
    CLRefreshParams *new_cl_refresh_params();
    void delete_cl_refresh_params(CLRefreshParams *cl_refresh_params);
    PGconn *clientlogic_pgconn_init(const char *user, size_t user_len, const char *database, size_t database_len,
        const char *application_name, size_t application_name_len, CursorInterface *cursor_interface,
        CursorDeallocator cursor_deallocaor, int enable_client_encryption);
    void clientlogic_pgconn_free(PGconn *pgconn);
    char *pgconn_errmsg(PGconn *pgconn);
    void pgconn_setserverversion(PGconn *pgconn, int server_version);
    void pgconn_setstdstrings(PGconn *pgconn, int std_strings);
    void clientlogic_pgconn_reset(PGconn *pgconn);
    PGClientLogic *get_client_logic(PGconn *conn);
    CursorInterface *get_new_driver_cursor(void *cursor, void (*destructor)(void *cursor),
        const char *(*oper)(void *cursor, int col), int (*load)(void *cursor, const char *query),
        int (*next)(void *cursor), int (*get_column_index)(void *cursor, const char *column_name),
        void (*clear_result)(void *cursor), int (*get_rows_number)(void *cursor), CursorDeallocator *r_destructor);
    int check_library();
    int get_last_refresh_cache_status(PGconn *conn, char **r_error);
    int clear_client_keys_c(PGconn *conn);
    int set_key_to_client_logic_c(PGconn *conn, size_t rsa_key_len, size_t ecdh_key_len, char *rsa_key, char *ecdh_key);
    int get_client_key_c(PGconn *conn, char **client_key, size_t *client_key_len);
    int fetched_columns_c(PGconn *conn);
    int get_cached_ceks_c(PGconn *conn, Oid **cek_oids, int **cek_algos, int **cek_buf_lens, char ***cek_bufs);

#ifdef __cplusplus
}
#endif // __cplusplus

#endif // LIB_FUNCS_MANAGER_H
