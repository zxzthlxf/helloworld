//go:build enable_ce
// +build enable_ce

#include <new>
#include "pgconn.h"
#include "go_cursor.h"
#include "lib_funcs_manager.h"
#include <assert.h>

PGconn *pgconn_init(void *go_conn, const char *user, int user_len,
                    const char *database, int database_len,
                    const char *application_name, int application_name_len, int enable_client_encryption)
{
    if (check_library() == 0) {
        return NULL;
    }
    CursorDeallocator cursor_deallocator;
    CursorInterface *cursor = GoCursor::make_cursor(go_conn, &cursor_deallocator);
    if (cursor == NULL)
    {
        return NULL;
    }
    return clientlogic_pgconn_init(user, user_len, database, database_len, application_name, application_name_len,
                                   cursor, cursor_deallocator, enable_client_encryption);
}

void pgconn_free(PGconn *pgconn)
{
    if (pgconn != NULL) {
        clientlogic_pgconn_free(pgconn);
    }
}

void pgconn_reset(PGconn *pgconn)
{
    clientlogic_pgconn_reset(pgconn);
}
