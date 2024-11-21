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
 * clientlogic_hooks.h
 *
 * IDENTIFICATION
 * 	  src/gitee.com/opengauss/openGauss-connector-go-pq/clientlogic_hooks.h
 *
 * -------------------------------------------------------------------------
 */
#ifndef HAVE_GO_CLIENTLOGIC_HOOKS_INCLUDE
#define HAVE_GO_CLIENTLOGIC_HOOKS_INCLUDE

#include <string.h>
#include "lib_funcs_manager.h"
typedef unsigned int Oid;

#ifdef __cplusplus
extern "C"
{
#endif // __cplusplus

typedef struct pg_conn PGconn;
typedef struct StatementData StatementData;
typedef struct CLRefreshParams CLRefreshParams;

int run_post_query(PGconn *conn, int rfq_cache_refresh_type, int conn_status);
StatementData *run_pre_query(PGconn *conn, const char *stmt_name, const char *query, int *client_side_error);
StatementData *run_pre_exec(PGconn *conn, const char *stmt_name, size_t params_count,
    const char * const * params_values, const int *params_lengths, const int *params_formats, int *client_side_error);
int accept_pending_statements(PGconn *conn);
char *deprocess_value(const char *processed_value, int processed_value_len, int original_type, size_t *length_output,
    int output_binary_format, PGconn *pgconn);
char *deprocess_record(const char *processed_value, int processed_value_len, int original_type, size_t *length_output,
    char *function_name, int output_binary_format, unsigned int type, PGconn *pgconn);

#ifdef __cplusplus
}
#endif // __cplusplus

#endif // HAVE_GO_CLIENTLOGIC_HOOKS_INCLUDE
