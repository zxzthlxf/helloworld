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
 * pgconn.h
 *
 * IDENTIFICATION
 * 	  src/gitee.com/opengauss/openGauss-connector-go-pq/pgconn.h
 *
 * -------------------------------------------------------------------------
 */
#ifndef HAVE_GO_PGCONN_INCLUDE
#define HAVE_GO_PGCONN_INCLUDE

#ifdef __cplusplus
extern "C" {
#endif // __cplusplus

typedef struct pg_conn PGconn;
PGconn *pgconn_init(void *go_conn, const char *user, int user_len, const char *database, int database_len,
    const char *application_name, int application_name_len, int enable_client_encryption);
void pgconn_free(PGconn *pgconn);
void pgconn_setserverversion(PGconn *pgconn, int server_version);
void pgconn_setstdstrings(PGconn *pgconn, int std_strings);
void pgconn_reset(PGconn *conn);


#ifdef __cplusplus
}
#endif // __cplusplus

#endif // HAVE_GO_PGCONN_INCLUDE
