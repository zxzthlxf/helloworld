//go:build enable_ce
// +build enable_ce

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
 * go_cursor.cpp
 *
 * IDENTIFICATION
 * 	  go_cursor.cpp
 *
 * -------------------------------------------------------------------------
 */

#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <new>
#include "go_cursor.h"
#include "_cgo_export.h"
#include "lib_funcs_manager.h"

/**
 * @brief static wrapper for get_col_value_by_idx
 *
 * @param cursor
 * @param column_index
 * @return const char*
 */
const char *GoCursor::static_get_col_value_by_idx(void *cursor, int column_index)
{
    return ((GoCursor *)cursor)->get_col_value_by_idx(column_index);
}

/**
 * @brief static wrapper for load
 *
 * @param cursor
 * @param query
 * @return int
 */
int GoCursor::static_load(void *cursor, const char *query)
{
    return ((GoCursor *)cursor)->load(query);
}

/**
 * @brief static wrapper for next
 *
 * @param cursor
 * @return int
 */
int GoCursor::static_next(void *cursor)
{
    return ((GoCursor *)cursor)->next();
}

/**
 * @brief static wrapper for get_column_index
 *
 * @param cursor
 * @param column_name
 * @return int
 */
int GoCursor::static_get_column_index(void *cursor, const char *column_name)
{
    return ((GoCursor *)cursor)->get_column_index(column_name);
}

/**
 * @brief static wrapper for clear_result
 *
 * @param cursor
 */
void GoCursor::static_clear_result(void *cursor)
{
    return ((GoCursor *)cursor)->clear_result();
}


/**
 * @brief static wrapper for get_rows_number
 *
 * @param cursor
 */
int GoCursor::static_get_rows_number(void *cursor)
{
    return ((GoCursor *)cursor)->get_rows_number();
}

/* *
 * Constructor
 */
GoCursor::GoCursor(void *go_conn)
    : m_go_conn(go_conn),
      m_resultset(NULL),
      m_columns_names(NULL),
      m_row_index(-1),
      m_rows_count(-1),
      m_columns_count(-1),
      m_error_code(0)
{
    m_driver_cursor = get_new_driver_cursor(this, GoCursorDeallocator, static_get_col_value_by_idx, static_load,
        static_next, static_get_column_index, static_clear_result, static_get_rows_number, &(m_driver_destructor));
}

/* *
 * Destrcutor
 */
GoCursor::~GoCursor()
{
    clear_result();
}

/**
 * @brief returns a new go cursor. Should be deallocated with delete.
 *
 * @param go_conn
 * @return CursorInterface*
 */
CursorInterface *GoCursor::make_cursor(void *go_conn, CursorDeallocator *cursor_deallocator)
{
    GoCursor *cursor = new (std::nothrow) GoCursor(go_conn);
    if (cursor == NULL){
        return NULL;
    }

    *cursor_deallocator = cursor->m_driver_destructor;
    return cursor->m_driver_cursor;
}

/* *
 * Loads the cursor data from libpq PGresult
 * @param[in] query query to use
 * @return on
 */
int GoCursor::load(const char *query)
{
    GoString gostring;
    gostring.p = query;
    gostring.n = strlen(query);
    Conncgo_SimpleQuery_return simple_query_return;
    simple_query_return = Conncgo_SimpleQuery(m_go_conn, gostring);
    m_resultset = simple_query_return.r0;
    m_columns_names = simple_query_return.r1;
    m_rows_count = simple_query_return.r2;
    m_columns_count = simple_query_return.r3;
    m_row_index = -1;
    return 1;
}

/* *
 * Move next record
 * @return true on success and false for end of cursor
 */
int GoCursor::next()
{
    if (m_row_index < m_rows_count - 1) {
        ++m_row_index;
        return 1;
    }
    return 0;
}

/* *
 * Retrieve column value from current record by index
 * @param[in] col column index
 * @return column value or null for invalid index
 */
const char *GoCursor::get_col_value_by_idx(int column_index) const
{
    if (m_row_index < 0 || m_row_index >= m_rows_count || column_index >= m_columns_count) {
        fprintf(stderr, "Client encryption operator[] failed with a bad index: %d", column_index);
        return NULL;
    }
    if (m_resultset != NULL) {
        return m_resultset[m_row_index][column_index];
    }
    return NULL;
}

/* *
 * Retrieve column value from current record by index
 * @param[in] col column index
 * @return column value or null for invalid index
 */
const char *GoCursor::operator[](int column_index) const
{
    return get_col_value_by_idx(column_index);
}

/* *
 * @param column_name column name
 * @return the column index of a given column name
 */
int GoCursor::get_column_index(const char *column_name) const
{
    for (int i = 0; i < m_columns_count; ++i) {
        if (strcasecmp(m_columns_names[i], column_name) == 0) {
            return i;
        }
    }
    return -1;
}

/* *
 * Clears the cursor results and release memory
 */
void GoCursor::clear_result()
{
    if (m_resultset != NULL) {
        for (int row_index = 0; row_index < m_rows_count; ++row_index) {
            for (int column_index = 0; column_index < m_columns_count; ++column_index) {
                free(m_resultset[row_index][column_index]);
                m_resultset[row_index][column_index] = NULL;
            }
            free(m_resultset[row_index]);
            m_resultset[row_index] = NULL;
        }
        free(m_resultset);
        m_resultset = NULL;
    }
    if (m_columns_names != NULL) {
        for (int i = 0; i < m_columns_count; ++i) {
            if (m_columns_names[i] != NULL) {
                free(m_columns_names[i]);
                m_columns_names[i] = NULL;
            }
        }

        free(m_columns_names);
        m_columns_names = NULL;
    }
    m_rows_count = 0;
    m_columns_count = 0;
}

/* *
 * get rows count, used when send token to TEE
 */
int GoCursor::get_rows_number()
{
    return m_rows_count;
}


void GoCursorDeallocator(void *cursor_interface)
{
    delete (GoCursor *)cursor_interface;
    cursor_interface = NULL;
}
