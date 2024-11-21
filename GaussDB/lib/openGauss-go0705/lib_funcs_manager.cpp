//go:build enable_ce
// +build enable_ce

#include <dlfcn.h>
#include <assert.h>
#include <mutex>
#include "lib_funcs_manager.h"

#define GETTER                                                \
    libpq_funcs_manager *getter = libpq_funcs_manager::get(); \
    if (getter == NULL) {                                     \
        assert(false);                                        \
    }

#define FUNC_SETTER(FUNCNAME)                                    \
    FUNCNAME = (func##FUNCNAME##func)dlsym(m_handle, #FUNCNAME); \
    if (FUNCNAME == NULL) {                                      \
        assert(FUNCNAME != NULL);                                \
        m_handle = NULL;                                         \
        return;                                                  \
    }

#define FUNC_DECL(FUNCNAME, ...)          \
    (*func##FUNCNAME##func)(__VA_ARGS__); \
    func##FUNCNAME##func FUNCNAME;

class libpq_funcs_manager {
public:
    static libpq_funcs_manager *get()
    {
        static libpq_funcs_manager mngr;
        // In case library was not found
        if (mngr.m_handle == NULL) {
            mngr.m_load_mutex.lock();   // Guard to avoid multiple allocs
            if (mngr.m_handle == NULL)  // For multithread, another thread might've already loaded it
            {
                mngr.load();
            }
            mngr.m_load_mutex.unlock();
        }
        if (mngr.m_handle == NULL) {
            return NULL; /* manager unuseable, so we return NULL */
        } else {
            return &mngr;
        }
    }

public:
    void *m_handle;
    /* usage is typdef <output param> FUNC_DECL(<function name in libpq>, <variadic function inputs>) */
    typedef const int *FUNC_DECL(get_rec_origial_ids, PGconn *, unsigned int, char const *);
    typedef size_t FUNC_DECL(get_rec_origial_ids_length, PGconn *, unsigned int, char const *);
    typedef int FUNC_DECL(de_process_record, PGconn *, char const *, unsigned long, int const *, unsigned long, int,
        unsigned char **, unsigned long *, int *);
    typedef void FUNC_DECL(deprocess_set_conn_c, Deprocess *, PGconn *);
    typedef int FUNC_DECL(get_deprocessed_c, Deprocess *deprocess, const unsigned char *processed_data,
        size_t processed_data_size, unsigned int type_oid, unsigned char **plain_text, size_t *plain_text_size);
    typedef int FUNC_DECL(processor_run_pre_query_c, StatementData *, int, int *);
    typedef void FUNC_DECL(delete_statementdata_c, StatementData *);
    typedef int FUNC_DECL(run_pre_exec_c, StatementData *);
    typedef int FUNC_DECL(accept_pending_statements_c, PGconn *, int);
    typedef int FUNC_DECL(free_mem_manager);
    typedef int FUNC_DECL(run_post_query_c, PGconn *, int);
    typedef Deprocess *FUNC_DECL(new_deprocess, int, int, int);
    typedef void FUNC_DECL(delete_deprocess, Deprocess *);
    typedef int FUNC_DECL(process_copy_chunk, PGconn *, const char *, int, char **);
    typedef int FUNC_DECL(is_any_refresh_cache_on_error, PGconn *);
    typedef int FUNC_DECL(pre_check_resend_query_on_error, PGconn *);
    typedef void FUNC_DECL(post_check_resend_query_on_error, PGconn *);
    typedef void FUNC_DECL(clientlogic_read_error, PGClientLogic *, const char, const char *, CLRefreshParams *);
    typedef int FUNC_DECL(deprocess_error_detail_c, PGconn *, const char *, char *);
    typedef StatementData *FUNC_DECL(new_statementdata_c, PGconn *, const char *, const char *, const size_t,
        const Oid *, const char * const *, const int *, const int *, int);
    typedef size_t FUNC_DECL(statement_data_get_params_count, StatementData *);
    typedef int FUNC_DECL(processed_query_pos_to_deprocessed_c, PGconn *, int);
    typedef const char *const *FUNC_DECL(statement_data_get_params_values, StatementData *);
    typedef int *FUNC_DECL(statement_data_get_params_lengths, StatementData *);
    typedef const char *FUNC_DECL(statement_data_get_query, StatementData *);
    typedef void FUNC_DECL(set_cl_rfq_cache_refresh_type, PGconn *, int);
    typedef int *FUNC_DECL(set_conn_status, PGconn *, int);
    typedef int FUNC_DECL(is_process_query, PGconn *);
    typedef CLRefreshParams *FUNC_DECL(new_cl_refresh_params);
    typedef void FUNC_DECL(delete_cl_refresh_params, CLRefreshParams *cl_refresh_params);
    typedef PGconn *FUNC_DECL(clientlogic_pgconn_init, const char *user, size_t user_len, const char *database,
        size_t database_len, const char *application_name, size_t application_name_len,
        CursorInterface *cursor_interface, CursorDeallocator cursor_deallocaor, int enable_client_encryption);
    typedef void FUNC_DECL(clientlogic_pgconn_free, PGconn *);
    typedef char *FUNC_DECL(pgconn_errmsg, PGconn *);
    typedef void FUNC_DECL(pgconn_setserverversion, PGconn *, int);
    typedef void FUNC_DECL(pgconn_setstdstrings, PGconn *, int);
    typedef void FUNC_DECL(clientlogic_pgconn_reset, PGconn *);
    typedef int FUNC_DECL(get_enum_val_by_string, const char *str);
    typedef PGClientLogic *FUNC_DECL(get_client_logic, PGconn *conn);
    typedef CursorInterface *FUNC_DECL(get_new_driver_cursor, void *cursor, void (*destructor)(void *cursor),
                                       const char *(*oper)(void *cursor, int col),
                                       int (*load)(void *cursor, const char *query), int (*next)(void *cursor),
                                       int (*get_column_index)(void *cursor, const char *column_name),
                                       void (*clear_result)(void *cursor), int (*get_rows_number)(void *cursor),
                                       CursorDeallocator* r_destructor);
    typedef int FUNC_DECL(get_last_refresh_cache_status, PGconn *conn, char **r_error);
    typedef int FUNC_DECL(clear_client_keys_c, PGconn *conn);
    typedef int FUNC_DECL(set_key_to_client_logic_c, PGconn *, size_t, size_t, char *, char *);
    typedef int FUNC_DECL(get_client_key_c, PGconn *, char **, size_t *);
    typedef int FUNC_DECL(fetched_columns_c, PGconn *);
    typedef int FUNC_DECL(get_cached_ceks_c, PGconn *, Oid **, int **, int **, char ***);
    
private :
    libpq_funcs_manager() : m_handle(NULL)
    {
        this->load();
        }
        ~libpq_funcs_manager()
        {
            dlclose(m_handle);
        }

        void load()
        {
#ifdef DEBUG_NOLIB
            return;
#endif
            assert(m_handle == NULL);
            m_handle = dlopen("libpq_ce.so", RTLD_LAZY);
            if (m_handle == NULL)  // Failed to open library
            {
                return;
            }
            FUNC_SETTER(get_rec_origial_ids);
            FUNC_SETTER(get_rec_origial_ids_length);
            FUNC_SETTER(de_process_record);
            FUNC_SETTER(deprocess_set_conn_c);
            FUNC_SETTER(get_deprocessed_c);
            FUNC_SETTER(processor_run_pre_query_c);
            FUNC_SETTER(delete_statementdata_c);
            FUNC_SETTER(run_pre_exec_c);
            FUNC_SETTER(free_mem_manager);
            FUNC_SETTER(accept_pending_statements_c);
            FUNC_SETTER(new_deprocess);
            FUNC_SETTER(delete_deprocess);
            FUNC_SETTER(run_post_query_c);
            FUNC_SETTER(process_copy_chunk);
            FUNC_SETTER(is_any_refresh_cache_on_error);
            FUNC_SETTER(pre_check_resend_query_on_error);
            FUNC_SETTER(post_check_resend_query_on_error);
            FUNC_SETTER(clientlogic_read_error);
            FUNC_SETTER(deprocess_error_detail_c);
            FUNC_SETTER(new_statementdata_c);
            FUNC_SETTER(statement_data_get_params_count);
            FUNC_SETTER(processed_query_pos_to_deprocessed_c);
            FUNC_SETTER(statement_data_get_params_values);
            FUNC_SETTER(statement_data_get_params_lengths);
            FUNC_SETTER(statement_data_get_query);
            FUNC_SETTER(set_cl_rfq_cache_refresh_type);
            FUNC_SETTER(set_conn_status);
            FUNC_SETTER(is_process_query);
            FUNC_SETTER(new_cl_refresh_params);
            FUNC_SETTER(delete_cl_refresh_params);
            FUNC_SETTER(clientlogic_pgconn_init);
            FUNC_SETTER(clientlogic_pgconn_free);
            FUNC_SETTER(pgconn_errmsg);
            FUNC_SETTER(pgconn_setserverversion);
            FUNC_SETTER(pgconn_setstdstrings);
            FUNC_SETTER(clientlogic_pgconn_reset);
            FUNC_SETTER(get_enum_val_by_string);
            FUNC_SETTER(get_client_logic);
            FUNC_SETTER(get_new_driver_cursor);
            FUNC_SETTER(get_last_refresh_cache_status);
            FUNC_SETTER(clear_client_keys_c);
            FUNC_SETTER(set_key_to_client_logic_c);
            FUNC_SETTER(get_client_key_c);
            FUNC_SETTER(fetched_columns_c);
            FUNC_SETTER(get_cached_ceks_c);
        }
        std::mutex m_load_mutex;
};

int check_library() {
    return libpq_funcs_manager::get() != NULL;
}

int run_post_query_c(PGconn *conn, int b)
{
    GETTER
    return getter->run_post_query_c(conn, b);
}

const int *get_rec_origial_ids(PGconn *conn, unsigned int a, char const *b)
{
    GETTER
    return getter->get_rec_origial_ids(conn, a, b);
}

size_t get_rec_origial_ids_length(PGconn *conn, unsigned int a, char const *b)
{
    GETTER
    return getter->get_rec_origial_ids_length(conn, a, b);
}

int de_process_record(PGconn *a, char const *b, unsigned long c, int const *d, unsigned long e, int f,
                      unsigned char **g, unsigned long *h, int *i)
{
    GETTER
    return getter->de_process_record(a, b, c, d, e, f, g, h, i);
}

Deprocess *new_deprocess(int a, int b, int c)
{
    GETTER
    return getter->new_deprocess(a, b, c);
}

void delete_deprocess(Deprocess *a)
{
    GETTER
    return getter->delete_deprocess(a);
}

void deprocess_set_conn_c(Deprocess *deprocess, PGconn *a)
{
    GETTER
    getter->deprocess_set_conn_c(deprocess, a);
}

int get_deprocessed_c(Deprocess *deprocess, const unsigned char *processed_data, size_t processed_data_size,
                      unsigned int type_oid, unsigned char **plain_text, size_t *plain_text_size)
{
    GETTER
    return getter->get_deprocessed_c(deprocess, processed_data, processed_data_size, type_oid, plain_text,
                                     plain_text_size);
}

int processor_run_pre_query_c(StatementData *a, int is_inner_query, int *failed_to_parse)
{
    GETTER
    return getter->processor_run_pre_query_c(a, is_inner_query, failed_to_parse);
}

int get_enum_val_by_string(const char *str)
{
    GETTER
    return getter->get_enum_val_by_string(str);
}

void delete_statementdata_c(StatementData *data)
{
    GETTER
    getter->delete_statementdata_c(data);
}

int run_pre_exec_c(StatementData *a)
{
    GETTER
    return getter->run_pre_exec_c(a);
}

int accept_pending_statements_c(PGconn *a, int b)
{
    GETTER
    return getter->accept_pending_statements_c(a, b);
}

int process_copy_chunk(PGconn *conn, const char *in_buffer, int msg_length, char **buffer)
{
    GETTER
    return getter->process_copy_chunk(conn, in_buffer, msg_length, buffer);
}

void post_check_resend_query_on_error(PGconn *conn)
{
    GETTER
    getter->post_check_resend_query_on_error(conn);
}

void free_mem_manager()
{
    GETTER
    getter->free_mem_manager();
}

int pre_check_resend_query_on_error(PGconn *conn)
{
    GETTER
    return getter->pre_check_resend_query_on_error(conn);
}

int is_any_refresh_cache_on_error(PGconn *conn)
{
    GETTER
    return getter->is_any_refresh_cache_on_error(conn);
}

void clientlogic_read_error(PGClientLogic *client_logic, const char id, const char *data,
                            CLRefreshParams *cl_refresh_params)
{
    GETTER
    getter->clientlogic_read_error(client_logic, id, data, cl_refresh_params);
}

int deprocess_error_detail_c(PGconn *conn, const char *value, char *contents)
{
    GETTER
    return getter->deprocess_error_detail_c(conn, value, contents);
}

StatementData *new_statementdata_c(PGconn *aconn, const char *astmt_name, const char *aquery, const size_t an_params,
    const Oid *aparam_types, const char * const * aparam_values, const int *aparam_lengths, const int *aparam_formats,
    int is_direct_stmt_flow)
{
    GETTER
    return getter->new_statementdata_c(aconn, astmt_name, aquery, an_params, aparam_types, aparam_values,
        aparam_lengths, aparam_formats, is_direct_stmt_flow);
}

size_t statement_data_get_params_count(StatementData *statement_data)
{
    GETTER
    return getter->statement_data_get_params_count(statement_data);
}

int processed_query_pos_to_deprocessed_c(PGconn *conn, int processed_pos)
{
    GETTER
    return getter->processed_query_pos_to_deprocessed_c(conn, processed_pos);
}

const char *const *statement_data_get_params_values(StatementData *statement_data)
{
    GETTER
    return getter->statement_data_get_params_values(statement_data);
}

int *statement_data_get_params_lengths(StatementData *statement_data)
{
    GETTER
    return getter->statement_data_get_params_lengths(statement_data);
}

const char *statement_data_get_query(StatementData *statement_data)
{
    GETTER
    return getter->statement_data_get_query(statement_data);
}

void set_cl_rfq_cache_refresh_type(PGconn *conn, int rfq)
{
    GETTER
    getter->set_cl_rfq_cache_refresh_type(conn, rfq);
}

void set_conn_status(PGconn *conn, int status)
{
    GETTER
    getter->set_conn_status(conn, status);
}

int is_process_query(PGconn *conn)
{
    GETTER
    return getter->is_process_query(conn);
}

CLRefreshParams *new_cl_refresh_params()
{
    GETTER
    return getter->new_cl_refresh_params();
}

void delete_cl_refresh_params(CLRefreshParams *cl_refresh_params)
{
    GETTER
    getter->delete_cl_refresh_params(cl_refresh_params);
}

PGconn *clientlogic_pgconn_init(const char *user, size_t user_len, const char *database, size_t database_len,
    const char *application_name, size_t application_name_len, CursorInterface *cursor_interface,
    CursorDeallocator cursor_deallocaor, int enable_client_encryption)
{
    GETTER
    return getter->clientlogic_pgconn_init(user, user_len, database, database_len, application_name,
        application_name_len, cursor_interface, cursor_deallocaor, enable_client_encryption);
}

void clientlogic_pgconn_free(PGconn *pgconn)
{
    GETTER
    getter->clientlogic_pgconn_free(pgconn);
}

char *pgconn_errmsg(PGconn *pgconn)
{
    GETTER
    return getter->pgconn_errmsg(pgconn);
}

void pgconn_setserverversion(PGconn *pgconn, int server_version)
{
    GETTER
    getter->pgconn_setserverversion(pgconn, server_version);
}

void pgconn_setstdstrings(PGconn *pgconn, int std_strings)
{
    GETTER
    getter->pgconn_setstdstrings(pgconn, std_strings);
}

void clientlogic_pgconn_reset(PGconn *pgconn)
{
    GETTER
    getter->clientlogic_pgconn_reset(pgconn);
}

PGClientLogic *get_client_logic(PGconn *conn)
{
    GETTER
    return getter->get_client_logic(conn);
}

CursorInterface *get_new_driver_cursor(void *cursor, void (*destructor)(void *cursor),
    const char *(*oper)(void *cursor, int col), int (*load)(void *cursor, const char *query), int (*next)(void *cursor),
    int (*get_column_index)(void *cursor, const char *column_name), void (*clear_result)(void *cursor),
    int (*get_rows_number)(void *cursor), CursorDeallocator *r_destructor)
{
    GETTER
    return getter->get_new_driver_cursor(cursor, destructor, oper, load, next, get_column_index, clear_result,
        get_rows_number, r_destructor);
}

int get_last_refresh_cache_status(PGconn *conn, char **r_error)
{
    GETTER
    return getter->get_last_refresh_cache_status(conn, r_error);
}

int clear_client_keys_c(PGconn *conn) {
    GETTER
    return getter->clear_client_keys_c(conn);
}

int set_key_to_client_logic_c(PGconn *conn, size_t rsa_key_len, size_t ecdh_key_len, char *rsa_key, char *ecdh_key) {
    GETTER
    return getter->set_key_to_client_logic_c(conn, rsa_key_len, ecdh_key_len, rsa_key, ecdh_key);
}

int get_client_key_c(PGconn *conn, char **client_key, size_t *client_key_len) {
    GETTER
    return getter->get_client_key_c(conn, client_key, client_key_len);
}

int fetched_columns_c(PGconn *conn) {
    GETTER
    return getter->fetched_columns_c(conn);
}

int get_cached_ceks_c(PGconn *conn, Oid **cek_oids, int **cek_algos, int **cek_buf_lens, char ***cek_bufs) {
    GETTER
    return getter->get_cached_ceks_c(conn, cek_oids, cek_algos, cek_buf_lens, cek_bufs);
}