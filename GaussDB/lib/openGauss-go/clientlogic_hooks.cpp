//go:build enable_ce
// +build enable_ce

#include "clientlogic_hooks.h"
#include "lib_funcs_manager.h"
#include <assert.h>

static StatementData *new_statementdata(PGconn *aconn, const char *astmt_name, const char *aquery)
{
    return new_statementdata_c(aconn, astmt_name, aquery, 0, 0, 0, 0, 0, 0);
}

static StatementData *new_statementdata(PGconn *aconn, const char *astmt_name, const size_t an_params,
                                        const Oid *aparam_types, const char *const *aparam_values,
                                        const int *aparam_lengths, const int *aparam_formats, int is_direct_stmt_flow)
{
    return new_statementdata_c(aconn, astmt_name, NULL, an_params, aparam_types, aparam_values, aparam_lengths,
                               aparam_formats, is_direct_stmt_flow);
}

/**
 * @brief wrapper function for the clientlogic "run_post_query" hook
 *
 * @param conn - pgqconn stub which the clientlogic in the libpq works with
 * @param rfq_cache_refresh_type - the "cache refresh type" as recieved by last RFQ token
 * @param conn_status - the status of the connection if it's connected or not.
 * @return int
 */
int run_post_query(PGconn *conn, int rfq_cache_refresh_type, int conn_status)
{
    set_cl_rfq_cache_refresh_type(conn, rfq_cache_refresh_type);
    set_conn_status(conn, conn_status);

    return run_post_query_c(conn, false) ? 1 : -1;
}

/**
 * @brief wrapper function for de_process_record
 *
 * @param processed_value  - string
 * @param original_type  - int
 * @return string
 */
char *deprocess_record(const char *processed_value, int processed_value_len, int original_type, size_t *length_output,
                       char *function_name, int output_binary_format, unsigned int typ, PGconn *pgconn)
{
    if (processed_value == NULL) {
        return (NULL);
    }

    const int *orig_ids = get_rec_origial_ids(pgconn, typ, function_name);

    if (orig_ids == NULL) {
        return NULL;
    }

    size_t length = get_rec_origial_ids_length(pgconn, typ, function_name);

    int is_value_decrypted = false;
    unsigned char *deprocessed = NULL;
    bool success = false;
    int format_type_binary = get_enum_val_by_string("Deprocess::FORMAT_TYPE_BINARY");
    int format_type_text = get_enum_val_by_string("Deprocess::FORMAT_TYPE_TEXT");

    if (de_process_record(pgconn, processed_value, processed_value_len, orig_ids, length,
                          (output_binary_format) ? format_type_binary : format_type_text, &deprocessed, length_output,
                          &is_value_decrypted)) {
        if (is_value_decrypted) {
            success = true;
        }
    }

    if (!success) {
        return NULL;
    }

    if (deprocessed == NULL) {
        return NULL;
    }
    return (char *)deprocessed;
}

/**
 * @brief wrapper function for deprocess.get_deprocessed
 *
 * @param processed_value  - string
 * @param original_type  - int
 * @return string
 */
char *deprocess_value(const char *processed_value, int processed_value_len, int original_type, size_t *length_output,
                      int output_binary_format, PGconn *pgconn)
{
    if (processed_value == NULL) {
        return (NULL);
    }

    unsigned char *deprocessed = NULL;
    int VALUE_RAW = get_enum_val_by_string("Deprocess::VALUE_RAW");
    int format_type_binary = get_enum_val_by_string("Deprocess::FORMAT_TYPE_BINARY");
    int format_type_text = get_enum_val_by_string("Deprocess::FORMAT_TYPE_TEXT");

    Deprocess *deprocess =
        new_deprocess(VALUE_RAW, format_type_binary, (output_binary_format) ? format_type_binary : format_type_text);
    if (deprocess == NULL) {
        return NULL;
    }
    deprocess_set_conn_c(deprocess, pgconn);

    if (get_deprocessed_c(deprocess, (const unsigned char *)processed_value, (size_t)processed_value_len, original_type,
                          &deprocessed, length_output) == 0) {
        delete_deprocess(deprocess);
        return NULL;
    }
    delete_deprocess(deprocess);

    if (deprocessed == NULL) {
        return NULL;
    }
    return (char *)deprocessed;
}

StatementData *run_pre_query(PGconn *conn, const char *stmt_name, const char *query, int *client_side_error)
{
    if (is_process_query(conn) == 0 || client_side_error == NULL) {
        return NULL;
    }

    *client_side_error = 0;

    StatementData *statement_data = new_statementdata(conn, stmt_name, query);
    if (statement_data == NULL) {
        return NULL;
    }

    if (!processor_run_pre_query_c(statement_data, 0, NULL)) {
        *client_side_error = 1;
        delete_statementdata_c(statement_data);
        return NULL;
    }

    return statement_data;
}

StatementData *run_pre_exec(PGconn *conn, const char *stmt_name, size_t params_count, const char *const *params_values,
                            const int *params_lengths, const int *params_formats, int *client_side_error)
{
    if (is_process_query(conn) == 0 || client_side_error == NULL) {
        return NULL;
    }
    *client_side_error = 0;

    StatementData *statement_data =
        new_statementdata(conn, stmt_name, params_count, NULL, params_values, params_lengths, params_formats, false);
    if (statement_data == NULL) {
        return NULL;
    }

    if (!run_pre_exec_c(statement_data)) {
        delete_statementdata_c(statement_data);
        *client_side_error = 1;
        return NULL;  // client err
    }

    return statement_data;
}

int accept_pending_statements(PGconn *conn)
{
    if (is_process_query(conn) == 0) {
        return 1;
    }

    return accept_pending_statements_c(conn, true) == 0 ? 1 : -1;
}