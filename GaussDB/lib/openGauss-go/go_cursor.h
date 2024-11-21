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
 * go_cursor.h
 *
 * IDENTIFICATION
 * 	  src/gitee.com/opengauss/openGauss-connector-go-pq/go_cursor.h
 *
 * -------------------------------------------------------------------------
 */
#ifndef HAVE_GO_CURSOR_INCLUDE
#define HAVE_GO_CURSOR_INCLUDE

#include <cstddef>
typedef struct CursorInterface CursorInterface;
typedef void (*CursorDeallocator)(CursorInterface *);

class GoCursor {
public:
    static CursorInterface *make_cursor(void *go_conn, CursorDeallocator *cursor_deallocator);
    ~GoCursor();
    const char *operator[](int col) const;
    int load(const char *query);

    int next();
    int get_column_index(const char *column_name) const;
    void clear_result();
    int get_rows_number();

private:
    GoCursor(void *go_conn);
    const char *get_col_value_by_idx(int column_index) const;
    static const char *static_get_col_value_by_idx(void *cursor, int column_index);
    static int static_load(void *cursor, const char *query);

    static int static_next(void *cursor);
    static int static_get_column_index(void *cursor, const char *column_name);
    static void static_clear_result(void *cursor);
    static int static_get_rows_number(void *cursor);
    void *m_go_conn = NULL;
    char ***m_resultset;
    char **m_columns_names;
    int m_row_index;
    int m_rows_count;
    int m_columns_count;
    int m_error_code;
    CursorDeallocator m_driver_destructor;
    CursorInterface *m_driver_cursor;
};

void GoCursorDeallocator(void *);

#endif /* HAVE_GO_CURSOR_INCLUDE */
