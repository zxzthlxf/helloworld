package pq

import (
	"bufio"
	"context"
	"crypto/md5"
	"crypto/tls"
	"database/sql/driver"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"gitee.com/opengauss/openGauss-connector-go-pq/oid"
)

// Common error types
var (
	ErrNotSupported              = errors.New("pq: Unsupported command")
	ErrInFailedTransaction       = errors.New("pq: Could not complete operation in a failed transaction")
	ErrSSLNotSupported           = errors.New("pq: SSL is not enabled on the server")
	ErrSSLKeyHasWorldPermissions = errors.New("pq: Private key file has group or world access. Permissions should be u=rw (0600) or less")
	ErrCouldNotDetectUsername    = errors.New("pq: Could not detect default username. Please provide one explicitly")

	errUnexpectedReady = errors.New("unexpected ReadyForQuery")
	errNoRowsAffected  = errors.New("no RowsAffected available after the empty statement")
	errNoLastInsertID  = errors.New("no LastInsertId available after the empty statement")
)

// Some special meta command
var (
	mSendToken  = "send_token"
	mClearToken = "clear_token"
)

/* TEE key info handlers */
const (
	teeInit      = 0
	teeNegotiate = 1
	teeDestroy   = 2
	teeSendKey   = 3
	teeCleanKey  = 4
)

type parameterStatus struct {
	// server version in the same format as server_version_num, or 0 if
	// unavailable
	serverVersion int

	// the current location based on the TimeZone value of the session, if
	// available
	currentLocation *time.Location
}

type transactionStatus byte

const (
	txnStatusIdle                transactionStatus = 'I'
	txnStatusIdleInTransaction   transactionStatus = 'T'
	txnStatusInFailedTransaction transactionStatus = 'E'
)

func (s transactionStatus) String() (string, error) {
	switch s {
	case txnStatusIdle:
		return "idle", nil
	case txnStatusIdleInTransaction:
		return "idle in transaction", nil
	case txnStatusInFailedTransaction:
		return "in a failed transaction", nil
	default:
		return "", fmt.Errorf("unknown transactionStatus %d", s)
	}
}

// connErr wraps driver.ErrBadConn for sql retry
type connErr struct {
	msg string
	err error
}

func (c connErr) Error() string {
	return fmt.Sprintf("%v: %v", c.msg, c.err)
}

func (c connErr) Unwrap() error {
	return c.err
}

type conn struct {
	c      net.Conn
	pgconn *PGconn        // currently used for the client encryption feature
	cn_ptr unsafe.Pointer // refers to the index in the pointers map (for cgo)
	buf    *bufio.Reader

	logger         Logger
	logLevel       LogLevel
	config         *Config
	fallbackConfig *FallbackConfig
	namei          int
	scratch        [512]byte
	txnStatus      transactionStatus
	txnFinish      func()
	// Save connection arguments to use during CancelRequest.
	dialer Dialer

	// Cancellation key data for use with CancelRequest messages.
	processID int
	secretKey int

	parameterStatus parameterStatus

	saveMessageType   byte
	saveMessageBuffer []byte

	// If true, this connection is bad and all public-facing functions should
	// return ErrBadConn.
	bad *atomic.Value

	// If set, this connection should never use the binary format when
	// receiving query results from prepared statements.  Only provided for
	// debugging.
	disablePreparedBinaryResult bool

	// Whether to always send []byte parameters over as binary.  Enables single
	// round-trip mode for non-prepared Query calls.
	binaryParameters bool

	// If true this connection is in the middle of a COPY
	inCopy                 bool
	isMasterForPreferSlave bool

	// If not nil, notices will be synchronously sent here
	noticeHandler func(*Error)

	// If not nil, notifications will be synchronously sent here
	notificationHandler func(*Notification)

	// mutex to safe-guard pgconn_free()
	pgconnMutex sync.RWMutex
}

func (cn *conn) LockReaderMutex() {
	if len(cn.config.EnableClientEncryption) != 0 {
		cn.pgconnMutex.RLock()
	}
}

func (cn *conn) UnlockReaderMutex() {
	if len(cn.config.EnableClientEncryption) != 0 {
		cn.pgconnMutex.RUnlock()
	}
}

func (cn *conn) LockWriterMutex() {
	if len(cn.config.EnableClientEncryption) != 0 {
		cn.pgconnMutex.Lock()
	}
}

func (cn *conn) UnlockWriterMutex() {
	if len(cn.config.EnableClientEncryption) != 0 {
		cn.pgconnMutex.Unlock()
	}
}

func (cn *conn) ResetSession(ctx context.Context) error {
	cn.LockReaderMutex()
	defer cn.UnlockReaderMutex()
	if cn.pgconn != nil {
		pgconn_reset(cn.pgconn)
	}
	return nil
}

func (cn *conn) shouldLog(lvl LogLevel) bool {
	return cn.logger != nil && cn.logLevel >= lvl
}

func (cn *conn) log(ctx context.Context, lvl LogLevel, msg string, data map[string]interface{}) {
	if !cn.shouldLog(lvl) {
		return
	}
	if data == nil {
		data = map[string]interface{}{}
	}
	if cn.c != nil && cn.processID != 0 {
		data["pid"] = cn.processID
	}

	cn.logger.Log(ctx, lvl, msg, data)
}

func (cn *conn) startTLS(tlsConfig *tls.Config) (err error) {
	if err = binary.Write(cn.c, binary.BigEndian, []int32{8, 80877103}); err != nil {
		return fmt.Errorf("cannot write binary: %w", err)
	}

	response := make([]byte, 1)
	if _, err = io.ReadFull(cn.c, response); err != nil {
		return connErr{
			msg: fmt.Sprintf("fail to read: %v", err),
			err: driver.ErrBadConn, // for database/sql errors.Is and retry
		}
	}

	if response[0] != 'S' {
		return ErrSSLNotSupported
	}

	cn.c = tls.Client(cn.c, tlsConfig)

	return nil
}

func (cn *conn) writeBuf(b byte) *writeBuf {
	cn.scratch[0] = b
	return &writeBuf{
		buf: cn.scratch[:5],
		pos: 1,
	}
}

type values map[string]string

func (cn *conn) isInTransaction() bool {
	return cn.txnStatus == txnStatusIdleInTransaction ||
		cn.txnStatus == txnStatusInFailedTransaction
}

func (cn *conn) setBad() {
	if cn.bad != nil {
		cn.bad.Store(true)
	}
}

func (cn *conn) getBad() bool {
	if cn.bad != nil {
		return cn.bad.Load().(bool)
	}
	return false
}

func (cn *conn) checkIsInTransaction(intxn bool) (err error) {
	if cn.isInTransaction() != intxn {
		cn.setBad()
		return fmt.Errorf("unexpected transaction status %v", cn.txnStatus)
	}
	return nil
}

func (cn *conn) Begin() (_ driver.Tx, err error) {
	cn.LockReaderMutex()
	defer cn.UnlockReaderMutex()
	return cn.begin("")
}

func (cn *conn) begin(mode string) (_ driver.Tx, err error) {
	if cn.getBad() {
		return nil, driver.ErrBadConn
	}

	if err = cn.checkIsInTransaction(false); err != nil {
		return nil, fmt.Errorf("cannot check is in transaction: %w", err)
	}
	_, commandTag, err := cn.simpleExec("BEGIN" + mode)
	if err != nil {
		return nil, fmt.Errorf("fail to simple exec: %w", err)
	}
	if commandTag != "BEGIN" {
		cn.setBad()
		return nil, fmt.Errorf("unexpected command tag %s", commandTag)
	}
	if cn.txnStatus != txnStatusIdleInTransaction {
		cn.setBad()
		return nil, fmt.Errorf("unexpected transaction status %v", cn.txnStatus)
	}
	return cn, nil
}

func (cn *conn) closeTxn() {
	if finish := cn.txnFinish; finish != nil {
		finish()
	}
}

func (cn *conn) Commit() (err error) {
	cn.LockReaderMutex()
	defer cn.UnlockReaderMutex()
	defer cn.closeTxn()
	if cn.getBad() {
		return driver.ErrBadConn
	}

	if err = cn.checkIsInTransaction(true); err != nil {
		return fmt.Errorf("cannot check is in transaction: %w", err)
	}
	// We don't want the client to think that everything is okay if it tries
	// to commit a failed transaction.  However, no matter what we return,
	// database/sql will release this connection back into the free connection
	// pool so we have to abort the current transaction here.  Note that you
	// would get the same behaviour if you issued a COMMIT in a failed
	// transaction, so it's also the least surprising thing to do here.
	if cn.txnStatus == txnStatusInFailedTransaction {
		if err := cn.rollback(); err != nil {
			return err
		}
		return ErrInFailedTransaction
	}

	_, commandTag, err := cn.simpleExec("COMMIT")
	if err != nil {
		if cn.isInTransaction() {
			cn.setBad()
		}
		return fmt.Errorf("fail to simple exec: %w", err)
	}
	if commandTag != "COMMIT" {
		cn.setBad()
		return fmt.Errorf("unexpected command tag %s", commandTag)
	}
	return cn.checkIsInTransaction(false)
}

func (cn *conn) Rollback() (err error) {
	cn.LockReaderMutex()
	defer cn.UnlockReaderMutex()
	defer cn.closeTxn()
	if cn.getBad() {
		return driver.ErrBadConn
	}
	return cn.rollback()
}

func (cn *conn) rollback() (err error) {
	if err = cn.checkIsInTransaction(true); err != nil {
		return fmt.Errorf("cannot check is in transaction: %w", err)
	}
	_, commandTag, err := cn.simpleExec("ROLLBACK")
	if err != nil {
		if cn.isInTransaction() {
			cn.setBad()
		}
		return fmt.Errorf("fail to simple exec: %w", err)
	}
	if commandTag != "ROLLBACK" {
		return fmt.Errorf("unexpected command tag %s", commandTag)
	}
	return cn.checkIsInTransaction(false)
}

func (cn *conn) gname() string {
	cn.namei++
	return strconv.FormatInt(int64(cn.namei), 10)
}

func (cn *conn) simpleExec(q string) (driver.Result, string, error) {
	if q == mSendToken {
		if cn.pgconn != nil && cn.config.EnableClientEncryption == "3" {
			err := cn.initEnclave()
			if err != nil {
				return nil, "", err
			} else {
				fmt.Println("Token cache enabled in Trusted Domain.")
				return driver.RowsAffected(0), "FETCH ", nil
			}
		}
	}

	if q == mClearToken {
		if cn.pgconn != nil && cn.config.EnableClientEncryption == "3" {
			err := cn.clearEnclave()
			if err != nil {
				return nil, "", err
			} else {
				fmt.Println("Token cache cleared in Trusted Domain.")
				return driver.RowsAffected(0), "FETCH ", nil
			}
		}
	}

	if cn.pgconn != nil {
		runtime.LockOSThread()
		var queryCstring *Cchar
		var err error

		q, queryCstring, err = cn.replaceQuery("", q)
		if err != nil {
			runtime.UnlockOSThread()
			return nil, "", fmt.Errorf("cannot replace query: %w", err)
		}
		defer Cfree(unsafe.Pointer(queryCstring))
		if err != nil {
			if cn.txnStatus == txnStatusIdleInTransaction {
				cn.txnStatus = txnStatusInFailedTransaction
			}
			runtime.UnlockOSThread()
			return nil, "", fmt.Errorf("cannot replace query: %w", err)
		}
	}
	b := cn.writeBuf('Q')
	b.string(q)
	if err := cn.send(b); err != nil {
		if cn.pgconn != nil {
			runtime.UnlockOSThread()
		}
		return nil, "", fmt.Errorf("fail to send: %w", err)
	}
	if cn.pgconn != nil {
		runtime.UnlockOSThread()
	}

	var (
		res    driver.Result
		cmdTag string
	)

	for {
		t, r, err := cn.recv1()
		if err != nil {
			cn.setBad()
			return nil, "", fmt.Errorf("cannot recv from conn: %w", err)
		}
		switch t {
		case 'C':
			s, err := r.string()
			if err != nil {
				return nil, "", fmt.Errorf("cannot get string from read buf: %w", err)
			}
			res, cmdTag, err = cn.parseComplete(s)
			if err != nil {
				return nil, "", fmt.Errorf("cannot parse complete: %w", err)
			}
		case 'E':
			err = parseError(r, cn)
			{
				var err error
				t, r, err = cn.recv1()
				if err != nil {
					cn.setBad()
					return nil, "", fmt.Errorf("cannot recv from conn: %w", err)
				}
			}
			fallthrough
		case 'Z':
			cn.processReadyForQuery(r)
			if err != nil {
				return nil, "", fmt.Errorf("got error from database: %w", err)
			}
			if res == nil && err == nil {
				return nil, "", errUnexpectedReady
			}
			// done
			return res, cmdTag, nil
		case 'I':
			res = emptyRows
		case 'T', 'D':
			// ignore any results
		default:
			cn.setBad()
			return nil, "", fmt.Errorf("unknown response for simple query: %q", t)
		}
	}
}

// replaceQuery allocate queryCstring (C String) that need to be used For server error handling. Must be free by caller
func (cn *conn) replaceQuery(stmtName string, query string) (new_query string, queryCstring *Cchar, err error) {
	stmtNameCstring := CString(stmtName)
	queryCstring = CString(query)
	var statement_data *StatementData
	client_side_err := 0
	statement_data = run_pre_query(cn.pgconn, stmtNameCstring, queryCstring, (*Cint)(unsafe.Pointer(&client_side_err)))
	new_query = query

	if statement_data != nil {
		new_query_cstring := statement_data_get_query(statement_data)
		if new_query_cstring != nil {
			new_query = GoString(new_query_cstring)
		}
	} else {
		if client_side_err != 0 {
			errstring := GoString(pgconn_errmsg(cn.pgconn))
			if errstring == "" {
				errstring = "Encountered a client error"
			}
			err = errors.New(errstring)
		}
	}
	if statement_data != nil {
		delete_statementdata_c(statement_data)
	}

	Cfree(unsafe.Pointer(stmtNameCstring))
	return new_query, queryCstring, err
}

func (cn *conn) simpleQuery(q string) (res *rows, err error) { // TODO: named return value
	if cn.pgconn != nil {
		var queryCstring *Cchar
		runtime.LockOSThread()
		q, queryCstring, err = cn.replaceQuery("", q)
		defer Cfree(unsafe.Pointer(queryCstring))
		if err != nil {
			runtime.UnlockOSThread()
			return nil, fmt.Errorf("cannot replace query: %w", err)
		}
	}

	b := cn.writeBuf('Q')
	b.string(q)
	if err = cn.send(b); err != nil {
		if cn.pgconn != nil {
			runtime.UnlockOSThread()
		}
		return nil, fmt.Errorf("fail to send: %w", err)
	}
	if cn.pgconn != nil {
		runtime.UnlockOSThread()
	}

	for {
		t, r, err := cn.recv1()
		if err != nil {
			cn.setBad()
			return nil, fmt.Errorf("unexpected message %q in simple query execution: %w", t, err)
		}
		switch t {
		case 'C', 'I': /* command complete or empty query */
			// We allow queries which don't return any results through Query as
			// well as Exec.  We still have to give database/sql a rows object
			// the user can close, though, to avoid connections from being
			// leaked.  A "rows" with done=true works fine for that purpose.
			if res == nil {
				res = &rows{
					cn: cn,
				}
			}
			// Set the result and tag to the last command complete if there wasn't a
			// query already run. Although queries usually return from here and cede
			// control to Next, a query with zero results does not.
			if t == 'C' { /* command complete */
				s, err := r.string()
				if err != nil {
					return nil, fmt.Errorf("cannot get string from read buf: %w", err)
				}
				res.result, res.tag, err = cn.parseComplete(s)
				if res.colNames != nil {
					return res, err
				}
			}
			res.done = true
		case 'E': /* error return */
			res, err = nil, parseError(r, cn)
			{
				var err error
				t, r, err = cn.recv1()
				if err != nil {
					cn.setBad()
					return nil, fmt.Errorf("unexpected message %q in simple query execution: %w", t, err)
				}
			}
			fallthrough
		case 'Z': /* backend is ready for new query */
			cn.processReadyForQuery(r)
			// done
			return res, err
		case 'D': /* Data Row */
			if res == nil {
				cn.setBad()
				return nil, errors.New("unexpected DataRow in simple query execution")
			}
			// the query didn't fail; kick off to Next
			if err = cn.saveMessage(t, r); err != nil {
				return nil, fmt.Errorf("cannot save message: %w", err)
			}
			return res, nil
		case 'T': /* Row Description */
			// res might be non-nil here if we received a previous
			// CommandComplete, but that's fine; just overwrite it
			res = &rows{cn: cn}

			des, err := parsePortalRowDescribe(r)
			if err != nil {
				return nil, fmt.Errorf("cannot parse protal row describe: %w", err)
			}
			res.rowsHeader = des

			// To work around a bug in QueryRow in Go 1.2 and earlier, wait
			// until the first DataRow has been received.
		default:
			cn.setBad()
			return nil, fmt.Errorf("unknown response for simple query: %q", t)
		}
	}
}

type noRows struct{}

var emptyRows noRows

var _ driver.Result = noRows{}

func (noRows) LastInsertId() (int64, error) {
	return 0, errNoLastInsertID
}

func (noRows) RowsAffected() (int64, error) {
	return 0, errNoRowsAffected
}

// Decides which column formats to use for a prepared statement.  The input is
// an array of type oids, one element per result column.
func decideColumnFormats(colTyps []fieldDesc, forceText bool) (colFmts []format, colFmtData []byte, err error) { // TODO: named return value
	if len(colTyps) == 0 {
		return nil, colFmtDataAllText, nil
	}

	colFmts = make([]format, len(colTyps))
	if forceText {
		return colFmts, colFmtDataAllText, nil
	}

	allBinary := true
	allText := true
	for i, t := range colTyps {
		switch t.OID {
		// This is the list of types to use binary mode for when receiving them
		// through a prepared statement.  If a type appears in this list, it
		// must also be implemented in binaryDecode in encode.go.
		case oid.T_byteawithoutorderwithequalcol:
			fallthrough
		case oid.T_byteawithoutordercol:
			fallthrough
		case oid.T__byteawithoutorderwithequalcol:
			fallthrough
		case oid.T__byteawithoutordercol:
			fallthrough
		case oid.T_bytea:
			fallthrough
		case oid.T_int8:
			fallthrough
		case oid.T_int4:
			fallthrough
		case oid.T_int2:
			fallthrough
		case oid.T_uuid:
			colFmts[i] = formatBinary
			allText = false

		default:
			allBinary = false
		}
	}

	if allBinary {
		return colFmts, colFmtDataAllBinary, nil
	} else if allText {
		return colFmts, colFmtDataAllText, nil
	} else {
		colFmtData = make([]byte, 2+len(colFmts)*2)
		binary.BigEndian.PutUint16(colFmtData, uint16(len(colFmts)))
		for i, v := range colFmts {
			binary.BigEndian.PutUint16(colFmtData[2+i*2:], uint16(v))
		}
		return colFmts, colFmtData, nil
	}
}

func (cn *conn) prepareTo(q, stmtName string) (st *stmt, err error) {
	st = &stmt{cn: cn, name: stmtName}

	if cn.pgconn != nil {
		var queryCstring *Cchar
		runtime.LockOSThread()
		q, queryCstring, err = cn.replaceQuery(stmtName, q)
		defer Cfree(unsafe.Pointer(queryCstring))
		if err != nil {
			runtime.UnlockOSThread()
			return nil, fmt.Errorf("cannot replace query: %w", err)
		}
	}

	// Prepare message
	b := cn.writeBuf('P')
	b.string(st.name)
	b.string(transferPlaceholder(q))

	/*
		the number of parameters the frontend wants to specifiy the data types for.
		since this is unused then the client must provide the values in the original data type of the column
	*/
	b.int16(0)

	// Describe message
	b.next('D')
	b.byte('S') // prepared statement
	b.string(st.name)

	// Sync early (before bind) so:
	// 1. we can get the response to the Describe
	// 2. handle errors here
	b.next('S')
	if err = cn.send(b); err != nil {
		if cn.pgconn != nil {
			runtime.UnlockOSThread()
		}
		return nil, fmt.Errorf("fail to send: %w", err)
	}
	if cn.pgconn != nil {
		runtime.UnlockOSThread()
	}

	if err = cn.readParseResponse(); err != nil {
		return nil, fmt.Errorf("cannot read parse response: %w", err)
	}

	st.paramTypes, st.colNames, st.colTyps, err = cn.readStatementDescribeResponse() // request and response info
	if err != nil {
		return nil, fmt.Errorf("cannot read statement describe response: %w", err)
	}
	st.colFmts, st.colFmtData, err = decideColumnFormats(st.colTyps, cn.disablePreparedBinaryResult) // response info only
	if err != nil {
		return nil, fmt.Errorf("cannot decide column formats %w", err)
	}
	if err = cn.readReadyForQuery(); err != nil {
		return nil, fmt.Errorf("cannot read ready for query: %w", err)
	}
	return st, nil
}

func transferPlaceholder(query string) string {
	var builder strings.Builder
	count := 1
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			builder.WriteString("$" + strconv.Itoa(count))
			count++
		} else {
			builder.WriteString(string(query[i]))
		}
	}
	return builder.String()
}

func (cn *conn) Prepare(q string) (_ driver.Stmt, err error) {
	cn.LockReaderMutex()
	defer cn.UnlockReaderMutex()
	return cn.prepare(q, true)
}

func (cn *conn) prepare(q string, check_resend bool) (stmt driver.Stmt, err error) {
	if cn.getBad() {
		return nil, driver.ErrBadConn
	}
	if check_resend {
		defer cn.prepareResend(q, &stmt, &err)
	}
	//defer cn.errRecover(&err)

	if len(q) >= 4 && strings.EqualFold(q[:4], "COPY") {
		s, err := cn.prepareCopyIn(q) // TODO: refactor error handle
		if err == nil {
			cn.inCopy = true
		}
		return s, err
	}

	st, err := cn.prepareTo(q, cn.gname())
	if err != nil {
		return nil, fmt.Errorf("fail to prepare to: %v", err) // return nil interface
	}

	return st, nil
}

/*
Resend function for conn.prepare.

	Check error if we encountered a cache-out-of-date error, and resend the query if needed
*/
func (cn *conn) prepareResend(q string, r_stmt *driver.Stmt, r_err *error) {
	if r_err == nil || *r_err == nil || r_stmt == nil || cn.pgconn == nil || is_any_refresh_cache_on_error(cn.pgconn) == false {
		return
	}

	defer post_check_resend_query_on_error(cn.pgconn) //May reset side effects from pre

	if pre_check_resend_query_on_error(cn.pgconn) == false {
		return
	}

	stmt, err := cn.prepare(q, false)
	if stmt != nil {
		if *r_stmt != nil {
			(*r_stmt).Close()
		}
		*r_stmt = stmt
	}
	*r_err = err
}

func pgconnNilSetter(pgconn **PGconn) {
	*pgconn = nil
}

func (cn *conn) Close() (err error) {
	cn.LockWriterMutex()
	defer cn.UnlockWriterMutex()
	// Ensure that cn.c.Close is always run. Since error handling is done with
	// cn.errRecover, the Close must be in a defer.
	defer cn.c.Close()
	defer func() {
		deletePointer(cn.cn_ptr)
		cn.cn_ptr = nil
	}()
	if cn.pgconn != nil {
		if cn.config.EnableAutoSendToken {
			err := cn.clearEnclave()
			if err != nil {
				fmt.Printf("Connection Close Warning: %s\n", err.Error())
			}
		}
		defer pgconnNilSetter(&cn.pgconn)
		defer pgconn_free(cn.pgconn)
	}
	// Don't go through send(); ListenerConn relies on us not scribbling on the
	// scratch buffer of this connection.
	return cn.sendSimpleMessage('X')
}

// Query Implement the "Queryer" interface
func (cn *conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	cn.LockReaderMutex()
	defer cn.UnlockReaderMutex()
	return cn.query(query, args, true)
}

func (cn *conn) query(query string, args []driver.Value, check_resend bool) (res *rows, err error) {
	if cn.getBad() {
		return nil, driver.ErrBadConn
	}
	if cn.inCopy {
		return nil, errCopyInProgress
	}
	if check_resend {
		defer cn.queryResend(query, args, &res, &err) // TODO: should not modify return value in defer
	}

	// Check to see if we can use the "simpleQuery" interface, which is
	// *much* faster than going through prepare/exec
	if len(args) == 0 {
		return cn.simpleQuery(query)
	}

	if cn.binaryParameters {
		if err = cn.sendBinaryModeQuery(query, args); err != nil {
			return nil, fmt.Errorf("cannot send binary mode query: %w", err)
		}

		if err = cn.readParseResponse(); err != nil {
			return nil, fmt.Errorf("cannot read parse response: %w", err)
		}

		if err = cn.readBindResponse(); err != nil {
			return nil, fmt.Errorf("cannot read bind response: %w", err)
		}
		res = &rows{cn: cn}
		res.rowsHeader, err = cn.readPortalDescribeResponse()
		if err != nil {
			return nil, fmt.Errorf("cannot read portal describe response: %w", err)
		}
		if err = cn.postExecuteWorkaround(); err != nil {
			return res, fmt.Errorf("cannot pose execute workaround: %w", err)
		}
		return res, nil
	}
	st, err := cn.prepareTo(query, "")
	if err != nil {
		return nil, fmt.Errorf("cannot prepare with query %s: %w", query, err)
	}

	if err = st.exec(args, true); err != nil {
		return nil, fmt.Errorf("cannot exec with value %v: %w", args, err)
	}

	return &rows{
		cn:         cn,
		rowsHeader: st.rowsHeader,
	}, nil
}

/*
Resend function for conn.query.

	Check error if we encountered a cache-out-of-date error, and resend the query if needed
*/
func (cn *conn) queryResend(query string, args []driver.Value, r_rows **rows, r_err *error) {
	if r_err == nil || *r_err == nil || r_rows == nil || cn.pgconn == nil || is_any_refresh_cache_on_error(cn.pgconn) == false {
		return
	}

	defer post_check_resend_query_on_error(cn.pgconn) //May reset side effects from pre
	if pre_check_resend_query_on_error(cn.pgconn) == false {
		return
	}

	rows, err := cn.query(query, args, false)
	if rows != nil {
		if *r_rows != nil {
			(*r_rows).Close()
		}
		*r_rows = rows
	}
	*r_err = err
}

func (cn *conn) Exec(query string, args []driver.Value) (res driver.Result, err error) {
	cn.LockReaderMutex()
	defer cn.UnlockReaderMutex()
	return cn.exec(query, args, true)
}

// Exec Implement the optional "Execer" interface for one-shot queries
func (cn *conn) exec(query string, args []driver.Value, check_resend bool) (res driver.Result, err error) {
	if cn.getBad() {
		return nil, driver.ErrBadConn
	}
	if check_resend {
		defer cn.execResend(query, args, &res, &err)
	}

	// Check to see if we can use the "simpleExec" interface, which is
	// *much* faster than going through prepare/exec
	if len(args) == 0 {
		// ignore commandTag, our caller doesn't care
		r, _, err := cn.simpleExec(query)
		if err != nil {
			return nil, fmt.Errorf("fail to simple exec: %w", err)
		}
		return r, nil
	}

	if cn.binaryParameters {
		if err := cn.sendBinaryModeQuery(query, args); err != nil {
			return nil, fmt.Errorf("cannot send binary mode query: %w", err)
		}

		if err = cn.readParseResponse(); err != nil {
			return nil, fmt.Errorf("cannot read parse response: %w", err)
		}
		if err = cn.readBindResponse(); err != nil {
			return nil, fmt.Errorf("cannot read bind response: %w", err)
		}
		if _, err := cn.readPortalDescribeResponse(); err != nil {
			return nil, fmt.Errorf("cannot read portal describe response: %w", err)
		}
		if err = cn.postExecuteWorkaround(); err != nil {
			return nil, fmt.Errorf("cannot read portal describe response: %w", err)
		}
		res, _, err = cn.readExecuteResponse("Execute")
		if err != nil {
			return nil, fmt.Errorf("cannot read execute response: %w", err)
		}

		return res, nil
	}
	// Use the unnamed statement to defer planning until bind
	// time, or else value-based selectivity estimates cannot be
	// used.
	st, err := cn.prepareTo(query, "")
	if err != nil {
		return nil, fmt.Errorf("cannot prepare query %s: %w", query, err)
	}
	r, err := st.Exec(args)
	if err != nil {
		return nil, fmt.Errorf("fail to exec: %w", err)
	}
	return r, err
}

/*
Resend function for conn.exec.

	Check error if we encountered a cache-out-of-date error, and resend the query if needed
*/
func (cn *conn) execResend(query string, args []driver.Value, r_res *driver.Result, r_err *error) {
	if r_err == nil || *r_err == nil || r_res == nil || cn.pgconn == nil || is_any_refresh_cache_on_error(cn.pgconn) == false {
		return
	}

	defer post_check_resend_query_on_error(cn.pgconn) //May reset side effects from pre

	if pre_check_resend_query_on_error(cn.pgconn) == false {
		return
	}

	res, err := cn.exec(query, args, false)
	if res != nil {
		*r_res = res
	}
	*r_err = err
}

type safeRetryError struct {
	msg string
	Err error
}

func (se *safeRetryError) Error() string {
	return se.Err.Error()
}

func (se *safeRetryError) Unwrap() error {
	return se.Err
}

func (cn *conn) send(m *writeBuf) error {
	n, err := cn.c.Write(m.wrap())
	if err != nil {
		if n == 0 || err == io.EOF {
			return &safeRetryError{Err: fmt.Errorf("fail to write %v: %w", err, driver.ErrBadConn)}
		}
		return &connErr{
			msg: fmt.Sprintf("fail to write %v", err),
			err: driver.ErrBadConn,
		}
	}

	if cn.pgconn != nil {
		free_mem_manager()
	}
	return nil
}

func (cn *conn) sendStartupPacket(m *writeBuf) error {
	if _, err := cn.c.Write((m.wrap())[1:]); err != nil {
		return connErr{
			msg: fmt.Sprintf("fail to write: %v", err),
			err: driver.ErrBadConn,
		}
	}
	return nil
}

// Send a message of type typ to the server on the other end of cn.  The
// message should have no payload.  This method does not use the scratch
// buffer.
func (cn *conn) sendSimpleMessage(typ byte) (err error) {
	if _, err = cn.c.Write([]byte{typ, '\x00', '\x00', '\x00', '\x04'}); err != nil {
		return connErr{
			msg: fmt.Sprintf("fail to write: %v", err),
			err: driver.ErrBadConn,
		}
	}
	return nil
}

// saveMessage memorizes a message and its buffer in the conn struct.
// recvMessage will then return these values on the next call to it.  This
// method is useful in cases where you have to see what the next message is
// going to be (e.g. to see whether it's an error or not) but you can't handle
// the message yourself.
func (cn *conn) saveMessage(typ byte, buf *readBuf) error {
	if cn.saveMessageType != 0 {
		cn.setBad()
		return fmt.Errorf("unexpected saveMessageType %d", cn.saveMessageType)
	}
	cn.saveMessageType = typ
	cn.saveMessageBuffer = *buf
	return nil
}

// recvMessage receives any message from the backend, or returns an error if
// a problem occurred while reading the message.
func (cn *conn) recvMessage(r *readBuf) (byte, error) {
	// workaround for a QueryRow bug, see exec
	if cn.saveMessageType != 0 {
		t := cn.saveMessageType
		*r = cn.saveMessageBuffer
		cn.saveMessageType = 0
		cn.saveMessageBuffer = nil
		return t, nil
	}

	x := cn.scratch[:5]
	if _, err := io.ReadFull(cn.buf, x); err != nil {
		return 0, connErr{
			msg: fmt.Sprintf("fail to read: %v", err),
			err: driver.ErrBadConn, // for database/sql errors.Is and retry
		}
	}

	// read the type and length of the message that follows
	t := x[0]
	n := int(binary.BigEndian.Uint32(x[1:])) - 4
	var y []byte
	if n <= len(cn.scratch) {
		y = cn.scratch[:n]
	} else {
		y = make([]byte, n)
	}
	if _, err := io.ReadFull(cn.buf, y); err != nil {
		return 0, connErr{
			msg: fmt.Sprintf("fail to read: %v", err),
			err: driver.ErrBadConn, // for database/sql errors.Is and retry
		}
	}
	*r = y
	return t, nil
}

// recv receives a message from the backend, but if an error happened while
// reading the message or the received message was an ErrorResponse, it returns err.
// NoticeResponses are ignored.  This function should generally be used only
// during the startup sequence.
func (cn *conn) recv() (t byte, r *readBuf, err error) {
	for {
		r = &readBuf{}
		t, err = cn.recvMessage(r)
		if err != nil {
			return 0, nil, fmt.Errorf("cannot recv message: %w", err)
		}
		switch t {
		case 'E':
			return 0, nil, parseError(r, cn)
		case 'N':
			if n := cn.noticeHandler; n != nil {
				n(parseError(r, cn))
			}
		case 'A':
			if n := cn.notificationHandler; n != nil {
				not, err := recvNotification(r)
				if err != nil {
					return 0, nil, fmt.Errorf("cannot recv notification: %w", err)
				}
				n(not)
			}
		default:
			return
		}
	}
}

// recv1Buf is exactly equivalent to recv1, except it uses a buffer supplied by
// the caller to avoid an allocation.
func (cn *conn) recv1Buf(r *readBuf) (byte, error) {
	for {
		t, err := cn.recvMessage(r)
		if err != nil {
			return 0, fmt.Errorf("cannot receive message: %w", err)
		}

		switch t {
		case 'A':
			if n := cn.notificationHandler; n != nil {
				not, err := recvNotification(r)
				if err != nil {
					return 0, fmt.Errorf("cannot recv notification: %w", err)
				}
				n(not)
			}
		case 'N':
			if n := cn.noticeHandler; n != nil {
				n(parseError(r, cn))
			}
		case 'S': // ParameterStatus
			if err := cn.processParameterStatus(r); err != nil {
				return 0, fmt.Errorf("cannot process parameter status: %w", err)
			}
		default:
			return t, nil
		}
	}
}

// recv1 receives a message from the backend
// while attempting to read it.  All asynchronous messages are ignored, with
// the exception of ErrorResponse.
func (cn *conn) recv1() (byte, *readBuf, error) {
	r := &readBuf{}
	t, err := cn.recv1Buf(r)
	return t, r, err
}

func (cn *conn) checkCertificate(tlsConn *tls.Conn) error {
	crlList := cn.config.crlList
	if crlList != nil && tlsConn != nil {
		for _, c := range tlsConn.ConnectionState().PeerCertificates {
			for _, revokeCert := range crlList.TBSCertList.RevokedCertificates {
				if fmt.Sprint(c.SerialNumber) == fmt.Sprint(revokeCert.SerialNumber) {
					return errors.New("the certificate has been revoked")
				}
			}
		}
	}

	return nil
}

func (cn *conn) startup() error {
	w := cn.writeBuf(0)
	w.int32(196659)
	// Send the backend the name of the database we want to connect to, and the
	// user we want to connect as.  Additionally, we send over any run-time
	// parameters potentially included in the connection string.  If the server
	// doesn't recognize any of them, it will reply with an error.

	var application_name string
	for k, v := range cn.config.RuntimeParams {
		w.string(k)
		w.string(v)
		if k == "application_name" {
			application_name = v
		}
	}
	if cn.config.Database != "" {
		w.string("database")
		w.string(cn.config.Database)
	}
	w.string("user")
	w.string(cn.config.User)

	if len(cn.config.EnableClientEncryption) != 0 {
		w.string("enable_full_encryption")
		user_cstring, user_clen := GetCString(&cn.config.User)
		database_cstring, database_clen := GetCString(&cn.config.Database)
		application_name_cstring, application_name_clen := GetCString(&application_name)
		/**
		initialize the pgconn object and pass the go "conn" object to it.
		also pass the connection string initial parameters
		*/
		res, err := strconv.Atoi(cn.config.EnableClientEncryption)
		if err != nil {
			return fmt.Errorf("cannot convey int from string: %w", err)
		}
		c_int := Cint(res)
		if (cn.cn_ptr == nil) || ((cn.cn_ptr != nil) &&
			(unsafe.Pointer(cn) != unsafe.Pointer(getPointer(cn.cn_ptr).(*conn)))) {
			cn.cn_ptr, err = addPointer(cn)
			if err != nil {
				return fmt.Errorf("cannot add pointer: %w", err)
			}
		}
		cn.pgconn = pgconn_init(cn.cn_ptr, user_cstring, user_clen,
			database_cstring, database_clen,
			application_name_cstring, application_name_clen, c_int)
		if cn.pgconn == nil {
			return errors.New("tried to use client logic and failed")
		}
		if len(cn.config.CryptoModuleInfo) != 0 {
			cCryptoModuleInfo := CString(cn.config.CryptoModuleInfo)
			isSuccess := set_crypto_module_info(cn.pgconn, cCryptoModuleInfo)
			Cfree(unsafe.Pointer(cCryptoModuleInfo))
			if !isSuccess {
				errString := GoString(pgconn_errmsg(cn.pgconn))
				if errString == "" {
					errString = "encountered a unreport client error"
				}
				return errors.New(errString)
			}
		}
		if len(cn.config.KeyInfo) != 0 {
			cKeyInfo := CString(cn.config.KeyInfo)
			isSuccess := set_key_info(cn.pgconn, cKeyInfo)
			Cfree(unsafe.Pointer(cKeyInfo))
			if !isSuccess {
				errString := GoString(pgconn_errmsg(cn.pgconn))
				if errString == "" {
					errString = "encountered a unreport client error"
				}
				return errors.New(errString)
			}
		}
		w.string(cn.config.EnableClientEncryption)
	} else {
		cn.pgconn = nil
	}
	w.string("")
	if err := cn.sendStartupPacket(w); err != nil {
		return fmt.Errorf("cannot send startup packet: %w", err)
	}

	tlsConn, ok := cn.c.(*tls.Conn)
	if ok {
		if err := cn.checkCertificate(tlsConn); err != nil {
			if err := cn.c.Close(); err != nil {
				return fmt.Errorf("cannot close connect: %w", err)
			}
			return fmt.Errorf("cannot check certificate: %w", err)
		}
	}

	for {
		t, r, err := cn.recv()
		if err != nil {
			return fmt.Errorf("cannot recv from conn: %w", err)
		}
		switch t {
		case 'K':
			cn.processBackendKeyData(r)
		case 'S':
			if err := cn.processParameterStatus(r); err != nil {
				return fmt.Errorf("cannot process parameter status: %w", err)
			}
		case 'R':
			if err := cn.auth(r); err != nil {
				return fmt.Errorf("fail to auth: %w", err)
			}
		case 'Z':
			cn.processReadyForQuery(r)
			found, err := cn.ValidateConnect()
			if err != nil {
				return fmt.Errorf("cannot validate connect: %w", err)
			}

			if cn.pgconn != nil && found == true && err == nil {
				var last_refresh_error_message *Cchar
				is_last_refresh_success := get_last_refresh_cache_status(cn.pgconn, (**Cchar)(unsafe.Pointer(&last_refresh_error_message)))
				if is_last_refresh_success == false {
					err_string := GoString(last_refresh_error_message)
					if last_refresh_error_message == nil || err_string == "" {
						err_string = "Failed to load cache for client logic feature."
					}
					if err = errors.New(err_string); err != nil {
						return fmt.Errorf("cannot convey to go string: %w", err)
					}
				}

				if cn.config.EnableAutoSendToken {
					err := cn.initEnclave()
					if err != nil {
						fmt.Printf("Connection Init Warning: %s", err.Error())
					}
				}
			}
			if err != nil {
				if err := cn.c.Close(); err != nil {
					return fmt.Errorf("cannot close connect: %w", err)
				}
			} else if found {
				return nil
			}
			return fmt.Errorf("ValidateConnect failed")
		default:
			return fmt.Errorf("unknown response for startup: %q", t)
		}
	}
}

func (cn *conn) auth(r *readBuf) error {
	var decodePwdByte []byte
	getPwdPlain := func() (string, error) { // TODO: refactor
		if len(cn.config.Password) == 0 {
			return "", errors.New("the server requested password-based authentication, but no password was provided")
		}
		var err error
		decodePwdByte, err = base64.StdEncoding.DecodeString(cn.config.Password)
		if err != nil {
			return "", fmt.Errorf("cannot decode string: %w", err)
		}
		return string(decodePwdByte), nil
	}

	switch code := r.int32(); code {
	case 0:
		// OK
	case 3:
		w := cn.writeBuf('p')

		plain, err := getPwdPlain()
		if err != nil {
			return fmt.Errorf("cannot get pwd plain: %w", err)
		}
		w.string(plain)
		if err = cn.send(w); err != nil {
			return fmt.Errorf("fail to send: %w", err)
		}

		t, r, err := cn.recv()
		if err != nil {
			return fmt.Errorf("cannot recv from conn: %w", err)
		}
		if t != 'R' {
			return fmt.Errorf("unexpected password response: %q", t)
		}

		if r.int32() != 0 {
			return fmt.Errorf("unexpected authentication response: %q", t)
		}
	case 5:
		s := string(r.next(4))
		w := cn.writeBuf('p')
		plain, err := getPwdPlain()
		if err != nil {
			return fmt.Errorf("cannot get pwd plain: %w", err)
		}
		w.string("md5" + md5s(md5s(plain+cn.config.User)+s))
		if err = cn.send(w); err != nil {
			return fmt.Errorf("fail to send: %w", err)
		}

		t, r, err := cn.recv()
		if err != nil {
			return fmt.Errorf("cannot recv from conn: %w", err)
		}
		if t != 'R' {
			return fmt.Errorf("unexpected password response: %q", t)
		}

		if r.int32() != 0 {
			return fmt.Errorf("unexpected authentication response: %q", t)
		}
	case 7: // GSSAPI, startup
		return fmt.Errorf("GSSAPI protocol not supported")
	case 8: // GSSAPI continue
		return fmt.Errorf("GSSAPI protocol not supported")

	case 10:
		passwordStoredMethod := r.int32()
		digest := ""
		if passwordStoredMethod == 0 || passwordStoredMethod == 2 {
			random64code := string(r.next(64))
			token := string(r.next(8))
			serverIteration := r.int32()
			plain, err := getPwdPlain()
			if err != nil {
				return fmt.Errorf("cannot get pwd plain: %w", err)
			}
			result := RFC5802Algorithm(plain, random64code, token, "", serverIteration)
			if len(result) == 0 {
				return fmt.Errorf("invalid username/password,login denied")
			}
			w := cn.writeBuf('p')
			w.buf = []byte("p")
			w.pos = 1
			w.int32(4 + len(result) + 1)
			w.bytes(result)
			w.byte(0)
			if err = cn.send(w); err != nil {
				return fmt.Errorf("fail to send: %w", err)
			}

			t, r, err := cn.recv()
			if err != nil {
				return fmt.Errorf("cannot recv from conn: %w", err)
			}

			if t != 'R' {
				return fmt.Errorf("unexpected password response: %q", t)
			}

			if r.int32() != 0 {
				return fmt.Errorf("unexpected authentication response: %q", t)
			}
		} else if passwordStoredMethod == 1 {
			s := string(r.next(4))
			plain, err := getPwdPlain()
			if err != nil {
				return fmt.Errorf("cannot get pwd plain: %w", err)
			}
			digest = "md5" + md5s(md5s(plain+cn.config.User)+s)
			w := cn.writeBuf('p')
			w.int16(4 + len(digest) + 1)
			w.string(digest)
			w.byte(0)
			if err = cn.send(w); err != nil {
				return fmt.Errorf("fail to send: %w", err)
			}
			t, r, err := cn.recv()
			if err != nil {
				return fmt.Errorf("cannot recv from conn: %w", err)
			}
			if t != 'R' {
				return fmt.Errorf("unexpected password response: %q", t)
			}

			if r.int32() != 0 {
				return fmt.Errorf("unexpected authentication response: %q", t)
			}
		} else {
			return fmt.Errorf("The  password-stored method is not supported ,must be plain , md5 or sha256.")
		}

	// AUTH_REQ_MD5_SHA256
	case 11:
		random64code := string(r.next(64))
		md5Salt := r.next(4)
		plain, err := getPwdPlain()
		if err != nil {
			return fmt.Errorf("cannot get pwd plain: %w", err)
		}
		result := Md5Sha256encode(plain, random64code, md5Salt)
		digest := []byte("md5")
		digest = append(digest, result...)
		w := cn.writeBuf('p')
		w.int32(4 + len(digest) + 1)
		w.bytes(digest)
		w.byte(0)
		if err = cn.send(w); err != nil {
			return fmt.Errorf("fail to send: %w", err)
		}

		t, r, err := cn.recv()
		if err != nil {
			return fmt.Errorf("cannot recv from conn: %w", err)
		}

		if t != 'R' {
			return fmt.Errorf("unexpected password response: %q", t)
		}

		if r.int32() != 0 {
			return fmt.Errorf("unexpected authentication response: %q", t)
		}

	default:
		return fmt.Errorf("unknown authentication response: %d", code)
	}
	clearBytes(decodePwdByte)
	return nil
}

func (cn *conn) ValidateConnect() (bool, error) {
	if cn.config.targetSessionAttrs == targetSessionAttrsAny {
		return true, nil
	}
	if cn.config.targetSessionAttrs == targetSessionAttrsReadWrite || cn.config.targetSessionAttrs == targetSessionAttrsReadOnly {
		sqlText := "show transaction_read_only"
		cn.log(context.Background(), LogLevelDebug, "Check server is transaction_read_only ?", map[string]interface{}{"sql": sqlText,
			"target_session_attrs": convertTargetSessionAttrToString(cn.config.targetSessionAttrs)})
		inReRows, err := cn.query(sqlText, nil, true)
		if err != nil {
			cn.log(context.Background(), LogLevelDebug, "err:"+err.Error(), map[string]interface{}{})
			return false, err
		}
		defer inReRows.Close()
		var dbTranReadOnly string
		lastCols := []driver.Value{&dbTranReadOnly}
		err = inReRows.Next(lastCols)
		if err != nil {
			cn.log(context.Background(), LogLevelDebug, "err:"+err.Error(), map[string]interface{}{})
			return false, err
		}
		readOnly := lastCols[0].(string)
		cn.log(context.Background(), LogLevelDebug, "Check server is readOnly ?", map[string]interface{}{"readOnly": readOnly})

		if cn.config.targetSessionAttrs == targetSessionAttrsReadWrite &&
			strings.EqualFold(readOnly, "off") {
			return true, nil
		} else if cn.config.targetSessionAttrs == targetSessionAttrsReadOnly &&
			strings.EqualFold(readOnly, "on") {
			return true, nil
		} else {
			return false, nil
		}
	} else {
		return cn.CheckConnectServerMasterSlave()
	}
}

func (cn *conn) CheckConnectServerMasterSlave() (bool, error) {
	sqlText := "select local_role,db_state from pg_stat_get_stream_replications()"
	cn.log(context.Background(), LogLevelDebug, "Check server is Master?", map[string]interface{}{"sql": sqlText,
		"target_session_attrs": convertTargetSessionAttrToString(cn.config.targetSessionAttrs)})
	inReRows, err := cn.query(sqlText, nil, true)
	if err != nil {
		cn.log(context.Background(), LogLevelDebug, "err:"+err.Error(), map[string]interface{}{})
		return false, err
	}
	defer inReRows.Close()
	var dbLocalRole, dbState string
	lastCols := []driver.Value{&dbLocalRole, &dbState}
	err = inReRows.Next(lastCols)
	if err != nil {
		cn.log(context.Background(), LogLevelDebug, "err:"+err.Error(), map[string]interface{}{})
		return false, err
	}
	localRole, ok := lastCols[0].(string)
	if !ok {
		return false, errors.New("expect return string")
	}
	state, ok := lastCols[1].(string)
	if !ok {
		return false, errors.New("expect return string")
	}
	cn.log(context.Background(), LogLevelDebug, "Check server is Master?", map[string]interface{}{"localRole": localRole,
		"dbState": state})
	isMaster := strings.EqualFold(localRole, "Primary") && strings.EqualFold(state, "Normal")
	if isMaster {
		if cn.config.targetSessionAttrs == targetSessionAttrsMaster {
			return true, nil
		} else if cn.config.targetSessionAttrs == targetSessionAttrsPreferSlave {
			cn.isMasterForPreferSlave = true
			return true, nil
		}
	} else if !isMaster && (cn.config.targetSessionAttrs == targetSessionAttrsSlave ||
		cn.config.targetSessionAttrs == targetSessionAttrsPreferSlave) {
		return true, nil
	}
	return false, nil
}

type format int

const formatText format = 0
const formatBinary format = 1

// One result-column format code with the value 1 (i.e. all binary).
var colFmtDataAllBinary = []byte{0, 1, 0, 1}

// No result-column format codes (i.e. all text).
var colFmtDataAllText = []byte{0, 0}

func IsColType(err error) bool {
	typeErr, ok := err.(colType)
	return ok && typeErr.ColType()
}

// colType implies a opaque error
type colType interface {
	ColType() bool
}

type colTypeErr struct {
	row int
	col int
}

func (c *colTypeErr) Error() string {
	return fmt.Sprintf("value in row %d,col %d type error", c.row, c.col)
}

func (c *colTypeErr) ColType() bool {
	return true
}

func checkColTypes(types []oid.Oid, v []driver.Value) error {
	tLen := len(types)
	if tLen == 0 {
		return fmt.Errorf("Div part exist zero.")
	}
	var colIdx int
	for i := 0; i < len(v); i++ {
		rowIdx := i / tLen
		oTy := findOid(v, i)
		idx := colIdx % tLen
		if !contains(oTy, types[idx]) {
			return &colTypeErr{rowIdx, idx}
		}
		colIdx++
	}
	return nil
}

// oids corresspond to parameter type
func findOid(v []driver.Value, idx int) []oid.Oid {
	boolOid := []oid.Oid{oid.T_bool, oid.T__bool, oid.T_char, oid.T__char, oid.T_byteawithoutorderwithequalcol,
		oid.T_byteawithoutordercol, oid.T__byteawithoutorderwithequalcol, oid.T__byteawithoutordercol, oid.T_varchar,
		oid.T__varchar, oid.T__nvarchar2, oid.T_nvarchar2, oid.T_bpchar, oid.T__bpchar, oid.T_text, oid.T__text}
	byteOid := []oid.Oid{oid.T_bytea, oid.T__bytea, oid.T_byteawithoutorderwithequalcol, oid.T_byteawithoutordercol,
		oid.T__byteawithoutorderwithequalcol, oid.T__byteawithoutordercol}
	stringOid := []oid.Oid{oid.T_bpchar, oid.T_raw, oid.T_blob, oid.T_clob, oid.T_text, oid.T_char, oid.T__raw,
		oid.T__char, oid.T__text, oid.T__bpchar, oid.T__varchar, oid.T_varchar, oid.T__blob, oid.T__clob,
		oid.T_unknown, oid.T_bit, oid.T__bit, oid.T_regtype, oid.T_regclass, oid.T__regtype, oid.T__regclass,
		oid.T__regproc, oid.T_regproc, oid.T__regoper, oid.T_regoper, oid.T_regoperator, oid.T__regoperator,
		oid.T__regprocedure, oid.T_regprocedure, oid.T_regdictionary, oid.T__regdictionary, oid.T_regconfig,
		oid.T__regconfig, oid.T_name, oid.T__name, oid.T_circle, oid.T__circle, oid.T__point, oid.T_box,
		oid.T__box, oid.T_lseg, oid.T_point, oid.T__lseg, oid.T_path, oid.T__path, oid.T__polygon, oid.T_polygon,
		oid.T_cidr, oid.T_inet, oid.T__inet, oid.T_macaddr, oid.T__macaddr, oid.T__tsvector, oid.T_tsvector,
		oid.T_tsquery, oid.T__tsquery, oid.T_uuid, oid.T__uuid, oid.T__json, oid.T_json, oid.T_hll, oid.T__hll,
		oid.T_int4range, oid.T_int8range, oid.T__int4range, oid.T__int8range, oid.T_daterange, oid.T__daterange,
		oid.T_numrange, oid.T__numrange, oid.T_tsrange, oid.T__tsrange, oid.T_tstzrange, oid.T__tstzrange,
		oid.T_hll_trans_type, oid.T__hll_trans_type, oid.T_tid, oid.T__tid, oid.T_xid, oid.T__xid, oid.T_cid,
		oid.T__cid, oid.T_xid32, oid.T__xid32, oid.T_oidvector_extend, oid.T_oidvector, oid.T__oidvector,
		oid.T_interval, oid.T__interval, oid.T_refcursor, oid.T_varbit, oid.T__varbit, oid.T_xml, oid.T__xml,
		oid.T_money, oid.T__money, oid.T_bytea, oid.T__bytea, oid.T_int1, oid.T_int2, oid.T_int4, oid.T_int8,
		oid.T_int16, oid.T_byteawithoutorderwithequalcol, oid.T_byteawithoutordercol, oid.T__nvarchar2, oid.T_nvarchar2,
		oid.T__byteawithoutorderwithequalcol, oid.T__byteawithoutordercol}
	intOid := []oid.Oid{oid.T_int1, oid.T_int2, oid.T_int4, oid.T_int8, oid.T_int16, oid.T_int2vector, oid.T_char,
		oid.T__char, oid.T_int2vector_extend, oid.T__int2vector_extend, oid.T__int2, oid.T__int2vector, oid.T__int4,
		oid.T__int8, oid.T__int16, oid.T_numeric, oid.T_varchar, oid.T__varchar, oid.T__nvarchar2, oid.T_nvarchar2,
		oid.T__int1, oid.T__numeric, oid.T_money, oid.T__money, oid.T_oid, oid.T__oid, oid.T_hll, oid.T__hll,
		oid.T_hash16, oid.T_hash32, oid.T__hash16, oid.T__hash32, oid.T_float4, oid.T_float8, oid.T__float4,
		oid.T__float8, oid.T_clob, oid.T_text, oid.T_byteawithoutorderwithequalcol, oid.T_byteawithoutordercol,
		oid.T__byteawithoutorderwithequalcol, oid.T__byteawithoutordercol, oid.T_bpchar, oid.T__bpchar, oid.T__text}
	floatOid := []oid.Oid{oid.T_float4, oid.T_float8, oid.T__float4, oid.T__float8, oid.T_clob, oid.T_numeric,
		oid.T__numeric, oid.T_money, oid.T__money, oid.T_hll, oid.T__hll, oid.T_text, oid.T_char, oid.T__char,
		oid.T_byteawithoutorderwithequalcol, oid.T_byteawithoutordercol, oid.T__byteawithoutorderwithequalcol,
		oid.T__byteawithoutordercol, oid.T_varchar, oid.T__varchar, oid.T__nvarchar2, oid.T_nvarchar2, oid.T_bpchar,
		oid.T__bpchar, oid.T__text}
	timeOid := []oid.Oid{oid.T_date, oid.T_time, oid.T_timestamp, oid.T__timestamp, oid.T__date, oid.T__time,
		oid.T_timestamptz, oid.T__timestamptz, oid.T_abstime, oid.T_reltime, oid.T_tinterval, oid.T__abstime,
		oid.T__reltime, oid.T__tinterval, oid.T_timetz, oid.T__timetz, oid.T_smalldatetime, oid.T__smalldatetime,
		oid.T_interval, oid.T__interval, oid.T_byteawithoutorderwithequalcol, oid.T_byteawithoutordercol,
		oid.T__byteawithoutorderwithequalcol, oid.T__byteawithoutordercol}
	switch v[idx].(type) {
	case bool:
		return boolOid
	case []byte:
		return byteOid
	case string:
		return stringOid
	case int64:
		return intOid
	case float64:
		return floatOid
	case time.Time:
		return timeOid
	default:
		return nil
	}
}

// convertParamValuesToC - convert params values list to C structure
//
//		@param paramsValues
//	 @return **Cchar
func convertParamValuesToC(paramsValues [][]byte, formats []int) (**Cchar, *Cint, *Cint) {

	// params values
	c_params_values := Cmalloc(Csize_t(len(paramsValues)) * Csize_t(unsafe.Sizeof(uintptr(0))))
	c_params_values_ref := (*[1<<30 - 1]*Cchar)(c_params_values)
	for i, paramValue := range paramsValues {
		if paramValue == nil {
			c_params_values_ref[i] = nil
		} else {
			c_params_values_ref[i] = CString((string)(paramValue))
		}
	}

	// params lengths
	c_params_lengths := Cmalloc(Csize_t(len(paramsValues)) * Csize_t(unsafe.Sizeof(int(0))))
	c_params_lengths_ref := (*[1<<30 - 1]Cint)(c_params_lengths)
	for i, paramValue := range paramsValues {
		if paramValue == nil {
			c_params_lengths_ref[i] = Cint(0)
		} else {
			c_params_lengths_ref[i] = Cint(len((string)(paramValue)))
		}
	}

	// params formats
	c_params_formats := Cmalloc(Csize_t(len(paramsValues)) * Csize_t(unsafe.Sizeof(int(0))))
	c_params_formats_ref := (*[1<<30 - 1]Cint)(c_params_formats)
	formats_len := len(formats)
	for i, _ := range paramsValues {
		c_format := Cint(0) // text
		if len(formats) != 0 && i < formats_len && formats[i] == 1 {
			c_format = Cint(1) // binary
		}
		c_params_formats_ref[i] = c_format
	}

	return (**Cchar)(c_params_values), (*Cint)(c_params_lengths), (*Cint)(c_params_formats)
}

func send_stmt_clientlogic_parameters(stmt_name string, pgconn *PGconn,
	writer *writeBuf, values [][]byte, formats []int) (err error) {

	c_params_values, c_params_lengths, c_params_formats := convertParamValuesToC(values, formats)
	c_stmt_name := CString(stmt_name)

	c_params_count := Csize_t(len(values))
	var statement_data *StatementData
	client_side_err := 0
	statement_data = run_pre_exec(pgconn, c_stmt_name, c_params_count, c_params_values, c_params_lengths, c_params_formats, (*Cint)(unsafe.Pointer(&client_side_err)))
	if statement_data != nil {
		params_count := statement_data_get_params_count(statement_data)
		params_values := statement_data_get_params_values(statement_data)
		params_lengths := statement_data_get_params_lengths(statement_data)

		params_values_ref := (*[1<<30 - 1]*Cchar)(unsafe.Pointer(params_values))
		params_lengths_ref := (*[1<<30 - 1]Cint)(unsafe.Pointer(params_lengths))
		var i Csize_t
		for i = 0; i < params_count; i++ {
			if params_values_ref[i] == nil {
				writer.int32(-1)
			} else {
				writer.int32((int)(params_lengths_ref[i]))
				writer.bytes(GoBytes(unsafe.Pointer(params_values_ref[i]), params_lengths_ref[i]))
			}
		}
		delete_statementdata_c(statement_data)
	}
	Cfree(unsafe.Pointer(c_stmt_name))
	for i, _ := range values {
		c_params_values_ref := (*[1<<30 - 1]*Cchar)(unsafe.Pointer(c_params_values))
		if c_params_values_ref[i] != nil {
			Cfree(unsafe.Pointer(c_params_values_ref[i]))
		}
	}
	Cfree(unsafe.Pointer(c_params_values))
	Cfree(unsafe.Pointer(c_params_lengths))
	Cfree(unsafe.Pointer(c_params_formats))

	if client_side_err != 0 {
		errstring := GoString(pgconn_errmsg(pgconn))
		if errstring == "" {
			errstring = "Encountered a client error"
		}
		err = errors.New(errstring)
	}
	return err
}

/*
*
this function is the myenv function that sends the client logic parameters.
the only reason why we have a separate function by the same name outside of the stmt
reciever is because of the very weird and wrong way BinaryParameters are implemented
in go pq. The BinaryParameters support should all be implemented in the stmt
reciever and not the way it is done today.
*/
func (st *stmt) send_clientlogic_parameters(pgconn *PGconn, writer *writeBuf,
	values [][]byte, formats []int) (err error) {
	return send_stmt_clientlogic_parameters(st.name, pgconn, writer, values, formats)
}

func (st *stmt) exec(v []driver.Value, check_retry bool) error {
	if st.cn.pgconn != nil && check_retry {
		defer st.exec_retry(v) //check & retry if we have client cache error
	}

	if len(v) == 0 && len(st.paramTypes) == 0 {
		err := st.execSingle(v, true)
		if err != nil {
			return fmt.Errorf("cannot exec: %w", err)
		}
	} else if len(v) != 0 && len(st.paramTypes) != 0 && len(v)%len(st.paramTypes) == 0 {
		// SELECT check argument size
		if len(v) != len(st.paramTypes) && st.colFmts != nil {
			return fmt.Errorf("got %d parameters but the statement may requires %d",
				len(v), len(st.paramTypes))
		}

		if len(v) == len(st.paramTypes) {
			err := st.execSingle(v, true)
			if err != nil {
				return fmt.Errorf("cannot exec single: %w", err)
			}
		} else {
			if st.cn.pgconn != nil && checkHaveCeCol(st.paramTypes) {
				return fmt.Errorf("got %d parameters but the statement may requires %d", len(v), len(st.paramTypes))
			}

			err := checkColTypes(st.paramTypes, v)
			if err != nil {
				return fmt.Errorf("colType error : %w", err)
			}
			err = st.execBatch(v)
			if err != nil {
				return fmt.Errorf("cannot exec batch: %w", err)
			}
		}
	} else {
		if len(st.paramTypes) == 0 {
			return fmt.Errorf("got %d parameters but the statement requires 0",
				len(v))
		}
		return fmt.Errorf("got %d parameters but the statement may requires %d",
			len(v), len(v)-len(v)%len(st.paramTypes)+len(st.paramTypes))
	}

	return nil
}

func checkHaveCeCol(types []oid.Oid) bool {
	for _, x := range types {
		switch x {
		case oid.T_byteawithoutorderwithequalcol:
			return true
		case oid.T_byteawithoutordercol:
			return true
		case oid.T__byteawithoutorderwithequalcol:
			return true
		case oid.T__byteawithoutordercol:
			return true
		default:
			return false
		}
	}
	return false
}

func (st *stmt) execSingle(v []driver.Value, check_retry bool) error {
	// validate Max args length
	if len(v) >= maxParamNum {
		return fmt.Errorf("got %d parameters but PostgreSQL only supports 65535 parameters", len(v))
	}

	cn := st.cn
	w := cn.writeBuf('B')
	w.byte(0)         // create a new unnamed portal
	w.string(st.name) // use the existing prepared statement

	if cn.binaryParameters {
		if err := cn.sendBinaryParameters(st.name, w, v); err != nil {
			return fmt.Errorf("cannot send binary parameters: %w", err)
		}
	} else {
		w.int16(0)      // magic number for "all text parameters" otherwise it should be the same as the number of total parameters
		w.int16(len(v)) // number of total parameters

		if cn.pgconn == nil {
			for i, x := range v {
				if x == nil {
					w.int32(-1)
				} else {
					b, err := encode(&cn.parameterStatus, x, st.paramTypes[i])
					if err != nil {
						return fmt.Errorf("cannot encode: %w", err)
					}
					w.int32(len(b))
					w.bytes(b)
				}
			}
		} else {
			var v_clientlogic_text [][]byte
			for i, x := range v {
				if x == nil {
					v_clientlogic_text = append(v_clientlogic_text, nil)
				} else {
					b, err := encode(&cn.parameterStatus, x, st.paramTypes[i])
					if err != nil {
						return fmt.Errorf("cannot encode: %w", err)
					}
					v_clientlogic_text = append(v_clientlogic_text, b)
				}
			}

			err := st.send_clientlogic_parameters(cn.pgconn, w, v_clientlogic_text, nil)
			if err != nil {
				return fmt.Errorf("cannot send client logic parameters: %w", err)
			}
		}
	}
	w.bytes(st.colFmtData)

	w.next('E') // EXECUTE
	w.byte(0)   // unnamed portal
	w.int32(0)  // unlimited number of rows

	w.next('S') // SYNC
	if err := cn.send(w); err != nil {
		return fmt.Errorf("fail to send: %w", err)
	}

	if err := cn.readBindResponse(); err != nil {
		return fmt.Errorf("cannot read bind response: %w", err)
	}
	return cn.postExecuteWorkaround()
}

func (st *stmt) exec_retry(v []driver.Value) error {
	if st.cn.pgconn == nil || is_any_refresh_cache_on_error(st.cn.pgconn) == false {
		return nil
	}

	if pre_check_resend_query_on_error(st.cn.pgconn) == false {
		if err := st.exec(v, false); err != nil {
			return fmt.Errorf("fail to exec: %w", err)
		}
	}
	post_check_resend_query_on_error(st.cn.pgconn) //May reset side effects from pre

	return nil
}

func (st *stmt) execBatch(v []driver.Value) error {
	types := len(st.paramTypes)
	if types == 0 {
		return fmt.Errorf("Div part exist zero.")
	}
	batchNum := len(v) / types

	err := checkPacketLength(v, st)
	if err != nil {
		return fmt.Errorf("cannot bind: %w", err)
	}

	cn := st.cn
	w := cn.writeBuf('U') // Bind
	w.int32(batchNum)     // batchNum
	w.byte(0)             // end of portal name
	w.string(st.name)     // statement name

	w.int16(types) // colNum

	// parameter format
	for _, paramType := range st.paramTypes {
		if isBinary(paramType) {
			w.int16(1)
		} else {
			w.int16(0)
		}
	}

	w.int16(0)
	w.int16(types)

	for i, x := range v {
		if x == nil {
			w.int32(-1)
		} else {
			var b []byte
			var err error
			if st.paramTypes[i%types] == oid.T_bytea {
				b, err = encode(&cn.parameterStatus, x, oid.T_char)
				if err != nil {
					return fmt.Errorf("cannot encode: %w", err)
				}
			} else {
				b, err = encode(&cn.parameterStatus, x, st.paramTypes[i%types])
				if err != nil {
					return fmt.Errorf("cannot encode: %w", err)
				}
			}
			w.int32(len(b)) // parameter length
			w.bytes(b)      // parameter value
		}
	}
	w.byte('E') // Execute
	w.byte(0)
	w.int32(0)

	w.next('S') // Sync
	if err := cn.send(w); err != nil {
		return fmt.Errorf("fail to send: %w", err)
	}

	if err := cn.readBindResponse(); err != nil {
		return fmt.Errorf("cannot read bind response: %w", err)
	}

	return cn.postExecuteWorkaround()
}

func checkPacketLength(v []driver.Value, st *stmt) error {
	var encodedSize int64
	types := len(st.paramTypes)

	for i, x := range v {
		if x == nil {
			encodedSize += 4 // length of int32
		} else {
			var b []byte
			var err error
			if st.paramTypes[i%types] == oid.T_bytea {
				b, err = encode(&st.cn.parameterStatus, x, oid.T_char)
			} else {
				b, err = encode(&st.cn.parameterStatus, x, st.paramTypes[i%types])
				if err != nil {
					return fmt.Errorf("cannot encode: %w", err)
				}
			}
			encodedSize += int64(4 + len(b)) // length of int32 + length of encoded parameter value
		}
	}

	/*
		encodedSize = (int32)packetLength + (int32)batchNum + (int8)end of portal name + len(statement name)
		+ (int8)end of statement name + (int16)len(st.paramTypes) + len(parameter format)*2 + (int16)0
		+ int(16)len(st.paramTypes) + (int8)'E' + (int8)end of portal name + int(32)row limit + encodedSize
	*/
	encodedSize = int64(4+4+1+len(st.name)+1+2+types*2+2+2+1+1+4) + encodedSize

	// limit the U packet length to 0x3fffffff
	if encodedSize > 0x3fffffff {
		return fmt.Errorf("bind message length %v too long. This can be caused by very large or incorrect "+
			"length specifications on InputStream parameters", encodedSize)
	}
	return nil
}

func isBinary(t oid.Oid) bool {
	types := []oid.Oid{oid.T_bytea, oid.T__bytea}
	return contains(types, t)
}

func contains(types []oid.Oid, val oid.Oid) bool {
	for _, t := range types {
		if t == val {
			return true
		}
	}
	return false
}

func (st *stmt) NumInput() int {
	return -1
}

// parseComplete parses the "command tag" from a CommandComplete message, and
// returns the number of rows affected (if applicable) and a string
// identifying only the command that was executed, e.g. "ALTER TABLE".  If the
// command tag could not be parsed, parseComplete returns error.
func (cn *conn) parseComplete(cmdTag string) (driver.Result, string, error) {
	commandsWithAffectedRows := []string{
		"SELECT ",
		// INSERT is handled below
		"UPDATE ",
		"DELETE ",
		"FETCH ",
		"MOVE ",
		"COPY ",
	}

	var affectedRows *string
	for _, tag := range commandsWithAffectedRows {
		if strings.HasPrefix(cmdTag, tag) {
			t := cmdTag[len(tag):]
			affectedRows = &t
			cmdTag = tag[:len(tag)-1]
			break
		}
	}
	// INSERT also includes the oid of the inserted row in its command tag.
	// Oids in user tables are deprecated, and the oid is only returned when
	// exactly one row is inserted, so it's unlikely to be of value to any
	// real-world application and we can ignore it.
	if affectedRows == nil && strings.HasPrefix(cmdTag, "INSERT ") {
		parts := strings.Split(cmdTag, " ")
		if len(parts) != 3 {
			cn.setBad()
			return nil, "", fmt.Errorf("unexpected INSERT command tag %s", cmdTag)
		}
		affectedRows = &parts[len(parts)-1]
		cmdTag = "INSERT"
	}
	// There should be no affected rows attached to the tag, just return it
	if affectedRows == nil {
		return driver.RowsAffected(0), cmdTag, nil
	}
	n, err := strconv.ParseInt(*affectedRows, 10, 64)
	if err != nil {
		cn.setBad()
		return nil, "", fmt.Errorf("could not parse commandTag: %w", err)
	}
	return driver.RowsAffected(n), cmdTag, nil
}

// QuoteIdentifier quotes an "identifier" (e.g. a table or a column name) to be
// used as part of an SQL statement.
// Any double quotes in name will be escaped.  The quoted identifier will be
// case sensitive when used in a query.  If the input string contains a zero
// byte, the result will be truncated immediately before it.
func QuoteIdentifier(name string) string {
	end := strings.IndexRune(name, 0)
	if end > -1 {
		name = name[:end]
	}
	return `"` + strings.Replace(name, `"`, `""`, -1) + `"`
}

func md5s(s string) string {
	h := md5.New()
	h.Write([]byte(s)) // TODO: unhandled error
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (cn *conn) sendBinaryParameters(stmt_name string, b *writeBuf, args []driver.Value) error {
	// Do one pass over the parameters to see if we're going to send any of
	// them over in binary.  If we are, create a paramFormats array at the
	// same time.
	var paramFormats []int
	for i, x := range args {
		_, ok := x.([]byte)
		if ok {
			if len(paramFormats) == 0 {
				paramFormats = make([]int, len(args))
			}
			paramFormats[i] = 1
		}
	}
	if len(paramFormats) == 0 {
		b.int16(0)
	} else {
		b.int16(len(paramFormats))
		for _, x := range paramFormats {
			b.int16(x)
		}
	}

	b.int16(len(args))

	if cn.pgconn == nil {
		for _, x := range args {
			if x == nil {
				b.int32(-1)
			} else {
				datum, err := binaryEncode(&cn.parameterStatus, x)
				if err != nil {
					return fmt.Errorf("fail to binary encode: %w", err)
				}
				b.int32(len(datum))
				b.bytes(datum)
			}
		}
	} else {
		var values_clientlogic [][]byte
		for _, x := range args {
			if x == nil {
				values_clientlogic = append(values_clientlogic, nil)
			} else {
				datum, err := binaryEncode(&cn.parameterStatus, x)
				if err != nil {
					return fmt.Errorf("fail to binary encode: %w", err)
				}
				values_clientlogic = append(values_clientlogic, datum)
			}
		}
		err := send_stmt_clientlogic_parameters(stmt_name, cn.pgconn, b, values_clientlogic, paramFormats)
		if err != nil {
			return fmt.Errorf("cannot send stmt client logic parameters: %w", err)
		}
	}

	return nil
}

func (cn *conn) sendBinaryModeQuery(q string, args []driver.Value) error {
	if len(args) >= 65536 {
		return fmt.Errorf("got %d parameters but PostgreSQL only supports 65535 parameters", len(args))
	}
	if cn.pgconn != nil {
		/*
			we ignore the error here because this function sendBinaryModeQuery()
			is not implemented correctly and it does not return an error.
			if an error occurred, the query string will be the original query
		*/
		var queryCstring *Cchar
		var err error
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		q, queryCstring, err = cn.replaceQuery("", q)
		defer Cfree(unsafe.Pointer(queryCstring))
		if err != nil {
			return fmt.Errorf("cannot replace query: %w", err)
		}
		// because SYNC is only sent in the end then it's all done in one go.
		accept_pending_statements(cn.pgconn)
	}

	b := cn.writeBuf('P')
	b.byte(0) // unnamed statement
	b.string(q)
	b.int16(0)

	b.next('B')
	b.int16(0) // unnamed portal and statement
	if err := cn.sendBinaryParameters("", b, args); err != nil {
		return fmt.Errorf("cannot send binary parameter: %w", err)
	}
	b.bytes(colFmtDataAllText)

	b.next('D')
	b.byte('P')
	b.byte(0) // unnamed portal

	b.next('E')
	b.byte(0)
	b.int32(0)

	b.next('S')
	return cn.send(b)
}

func (cn *conn) processParameterStatus(r *readBuf) error {
	param, err := r.string()
	if err != nil {
		return fmt.Errorf("cannot get string from read buf: %w", err)
	}
	val, err := r.string()
	if err != nil {
		return fmt.Errorf("cannot get string from read buf: %w", err)
	}

	cn.log(context.Background(), LogLevelInfo, "[connection parameter]", map[string]interface{}{param + ":": val})
	switch param {
	case "server_version":
		var major1 int
		var major2 int
		var minor int
		_, err = fmt.Sscanf(val, "%d.%d.%d", &major1, &major2, &minor)
		if err == nil {
			cn.parameterStatus.serverVersion = major1*10000 + major2*100
		}

		if cn.pgconn != nil {
			server_version := (major1 * 10000) + (major2 * 100) + minor
			pgconn_setserverversion(cn.pgconn, (Cint)(server_version))
		}

	case "TimeZone":
		cn.parameterStatus.currentLocation, err = time.LoadLocation(val)
		if err != nil {
			cn.parameterStatus.currentLocation = nil
		}

	case "standard_conforming_strings":
		if cn.pgconn != nil {
			value_int := 0
			if val == "on" {
				value_int = 1
			}
			pgconn_setstdstrings(cn.pgconn, (Cint)(value_int))
		}
	default:
		// ignore
	}

	return nil
}

func (cn *conn) processReadyForQuery(r *readBuf) {
	cn.txnStatus = transactionStatus(r.byte())
	/* if the pgconn is initialized, we can assume the client logic was turned on */
	if cn.pgconn != nil {
		/**
		rfq_cache_refresh_type is dereived from this token
		but unlike in libpq, the conn status is always "estasblished" (0) at this point.
		*/
		rfq_cache_refresh_type := int(r.byte())
		run_post_query(cn.pgconn, Cint(rfq_cache_refresh_type), false)
	}
}

func (cn *conn) readReadyForQuery() error {
	t, r, err := cn.recv1()
	if err != nil {
		cn.setBad() // TODO: fetch with closure
		return fmt.Errorf("cannot recv from conn: %w", err)
	}
	switch t {
	case 'Z':
		cn.processReadyForQuery(r)
		return nil
	default:
		cn.setBad()
		return fmt.Errorf("unexpected message %q; expected ReadyForQuery", t)
	}
}

func (cn *conn) processBackendKeyData(r *readBuf) {
	cn.processID = r.int32()
	cn.secretKey = r.int32()
}

func (cn *conn) readParseResponse() error {
	t, r, err := cn.recv1()
	if err != nil {
		cn.setBad()
		return fmt.Errorf("cannot recv from conn: %w", err)
	}
	switch t {
	case '1': // Parse Complete
		return nil
	case 'E': // Error response
		err = parseError(r, cn)
		if err := cn.readReadyForQuery(); err != nil {
			return fmt.Errorf("cannot read ready for query: %w", err)
		}
		if err != nil {
			return fmt.Errorf("got error from database: %w", err)
		}
	default:
		cn.setBad()
		return fmt.Errorf("unexpected Parse response %q", t)
	}

	return errors.New("not reached")
}

/*
*
retrieve information about the current prepared statement
1. the data types of the parameters ($X) in the query
2. the names of the columns anticipating in the response
3. the data types of the columns anticipating in the response
*/
func (cn *conn) readStatementDescribeResponse() (paramTyps []oid.Oid, colNames []string, colTyps []fieldDesc, err error) {
	for {
		t, r, err := cn.recv1()
		if err != nil {
			cn.setBad()
			return nil, nil, nil, fmt.Errorf("cannot recv from conn: %w", err)
		}
		switch t {
		case 't': // ParameterDescription
			nparams := r.int16()
			paramTyps = make([]oid.Oid, nparams)
			for i := range paramTyps {
				paramTyps[i] = r.oid()
			}
		case 'n': // NoData
			return paramTyps, nil, nil, nil
		case 'T': // RowDescription
			colNames, colTyps, err = parseStatementRowDescribe(r)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("cannot parse statement row desceibe: %w", err)
			}
			return paramTyps, colNames, colTyps, nil
		case 'E': // ErrorResponse
			err = parseError(r, cn)
			if err := cn.readReadyForQuery(); err != nil {
				return nil, nil, nil, fmt.Errorf("cannot read ready for query: %w", err)
			}
			if err != nil {
				return nil, nil, nil, fmt.Errorf("got error from database: %w", err)
			}
		default:
			cn.setBad()
			return nil, nil, nil, fmt.Errorf("unexpected Describe statement response %q", t)
		}
	}
}

func (cn *conn) readPortalDescribeResponse() (rowsHeader, error) {
	t, r, err := cn.recv1()
	if err != nil {
		cn.setBad()
		return rowsHeader{}, fmt.Errorf("cannot recv from conn: %w", err)
	}
	switch t {
	case 'T':
		return parsePortalRowDescribe(r)
	case 'n':
		return rowsHeader{}, nil
	case 'E':
		err = parseError(r, cn)
		if err := cn.readReadyForQuery(); err != nil {
			return rowsHeader{}, fmt.Errorf("cannot read ready for query: %w", err)
		}
		if err != nil {
			return rowsHeader{}, fmt.Errorf("got error from database: %w", err)
		}
	default:
		cn.setBad()
		return rowsHeader{}, fmt.Errorf("unexpected Describe response %q", t)
	}
	return rowsHeader{}, errors.New("not reached")
}

func (cn *conn) readBindResponse() error {
	t, r, err := cn.recv1()
	if err != nil {
		cn.setBad()
		return fmt.Errorf("cannot recv from conn: %w", err)
	}
	switch t {
	case '2': // BindComplete
		return nil
	case 'E': // ErrorResponse
		err = parseError(r, cn)
		if err := cn.readReadyForQuery(); err != nil {
			return fmt.Errorf("cannot read for query: %w", err)
		}
		if err != nil {
			return fmt.Errorf("got error from database: %w", err)
		}
	default:
		cn.setBad()
		return fmt.Errorf("unexpected Bind response %q", t)
	}

	return nil
}

func (cn *conn) postExecuteWorkaround() error {
	// Work around a bug in sql.DB.QueryRow: in Go 1.2 and earlier it ignores
	// any errors from rows.Next, which masks errors that happened during the
	// execution of the query.  To avoid the problem in common cases, we wait
	// here for one more message from the database.  If it's not an error the
	// query will likely succeed (or perhaps has already, if it's a
	// CommandComplete), so we push the message into the conn struct; recv1
	// will return it as the next message for rows.Next or rows.Close.
	// However, if it's an error, we wait until ReadyForQuery and then return
	// the error to our caller.
	for {
		t, r, err := cn.recv1()
		if err != nil {
			cn.setBad()
			return fmt.Errorf("cannot recv from conn: %w", err)
		}
		switch t {
		case 'E':
			err = parseError(r, cn)
			if err := cn.readReadyForQuery(); err != nil {
				return fmt.Errorf("cannot read ready for query: %w", err)
			}
			if err != nil {
				cn.setBad()
				return fmt.Errorf("unexpected CommandComplete after error: %w", err)
			}
		case 'C', 'D', 'I':
			// the query didn't fail, but we can't process this message
			if err = cn.saveMessage(t, r); err != nil {
				return fmt.Errorf("cannot save message: %w", err)
			}
			return nil
		default:
			cn.setBad()
			return fmt.Errorf("unexpected message during extended query execution: %q", t)
		}
	}
}

// Only for Exec(), since we ignore the returned data
func (cn *conn) readExecuteResponse(protocolState string) (res driver.Result, cmdTag string, err error) { // TODO: return unamed
	for {
		t, r, err := cn.recv1()
		if err != nil {
			cn.setBad()
			return nil, "", fmt.Errorf("unexpected CommandComplete after error: %w", err)
		}
		switch t {
		case 'C':
			s, err := r.string()
			if err != nil {
				return nil, "", fmt.Errorf("cannot get string from read buf: %w", err)
			}
			res, cmdTag, err = cn.parseComplete(s)
			if err != nil {
				return nil, "", fmt.Errorf("cannot parse complete: %w", err)
			}
		case 'Z':
			cn.processReadyForQuery(r)
			if res == nil && err == nil {
				return nil, "", errUnexpectedReady
			}
			return res, cmdTag, nil
		case 'E':
			err = parseError(r, cn)
		case 'T', 'D', 'I':
			if t == 'I' {
				res = emptyRows
			}
			// ignore any results
		default:
			cn.setBad()
			return nil, "", fmt.Errorf("unknown %s response: %q", protocolState, t)
		}
	}
}

func parseStatementRowDescribe(r *readBuf) (colNames []string, colTyps []fieldDesc, err error) {
	n := r.int16()
	colNames = make([]string, n)
	colTyps = make([]fieldDesc, n)
	for i := range colNames {
		s, err := r.string()
		if err != nil {
			return nil, nil, fmt.Errorf("cannot get string from read buf: %w", err)
		}
		colNames[i] = s
		r.next(6)
		colTyps[i].OID = r.oid()
		colTyps[i].Len = r.int16()
		colTyps[i].Mod = r.int32()
		// format code not known when describing a statement; always 0
		r.next(2)
	}
	return
}

func parsePortalRowDescribe(r *readBuf) (rowsHeader, error) {
	n := r.int16()
	colNames := make([]string, n)
	colFmts := make([]format, n)
	colTyps := make([]fieldDesc, n)
	for i := range colNames {
		s, err := r.string()
		if err != nil {
			return rowsHeader{}, fmt.Errorf("cannot get string from read buf: %w", err)
		}
		colNames[i] = s
		r.next(6)
		colTyps[i].OID = r.oid()
		colTyps[i].Len = r.int16()
		colTyps[i].Mod = r.int32()
		colFmts[i] = format(r.int16())
	}
	return rowsHeader{
		colNames: colNames,
		colFmts:  colFmts,
		colTyps:  colTyps,
	}, nil
}

// establish secure channel, send ceks to trusted domain safely, and destroy the secure channel
func (cn *conn) initEnclave() error {
	// get rsa public key and ecdh key from trusted domain, and generate client ecdh key (Packet V0)
	clientTokenLen, clientToken, initErr := cn.sendInitEcdhSecureChannel(teeInit, 0, nil)

	if initErr == nil {
		// send client ecdh key to trusted domain (Packet V1)
		_, _, initErr = cn.sendInitEcdhSecureChannel(teeNegotiate, clientTokenLen, clientToken)
	}

	if initErr == nil {
		cekOids, cekAlgos, cekBufLens, cekBufs, err := cn.fetchAndDecryptCek()
		if err == nil && cekOids != nil {
			for i := 0; i < len(cekOids); i++ {
				// send every cek to trusted domain (Packet V3)
				initErr = cn.sendOrDestroyCekInfo(teeSendKey, cekOids[i], cekAlgos[i], cekBufLens[i], cekBufs[i])
				if initErr != nil {
					break
				}
			}
		} else {
			initErr = err
		}
	}

	if initErr != nil {
		initErr = fmt.Errorf("can't finish cache token in Trusted Domain: %w", initErr)
	}

	clear_client_keys(cn.pgconn)

	// clear exchange information in trusted domain (Packet V2)
	_, _, err := cn.sendInitEcdhSecureChannel(teeDestroy, 0, nil)
	if initErr == nil && err != nil {
		return initErr
	}

	return initErr
}

// clear all cached ceks in trusted domain (Packet V4)
func (cn *conn) clearEnclave() error {
	err := cn.sendOrDestroyCekInfo(teeCleanKey, 0, 0, 0, nil)
	if err != nil {
		return fmt.Errorf("Failed to clear token in Trusted Domain: %w", err)
	}
	return nil
}

func (cn *conn) sendInitEcdhSecureChannel(sendType int, clientKeyLen int, clientKey []byte) (int, []byte, error) {
	b := cn.writeBuf('V')
	if sendType == teeNegotiate {
		b.int32(sendType)
		b.int32(clientKeyLen)
		b.bytes(clientKey)
	} else {
		b.int32(sendType)
	}

	if err := cn.send(b); err != nil {
		return 0, nil, err
	}

	resClientKeyLen, resClientKey, err := cn.readResponseForEnclave()
	return resClientKeyLen, resClientKey, err
}

func (cn *conn) sendOrDestroyCekInfo(opType int, cekOid uint, cekAlgo int, cekLen int, cek []byte) error {
	b := cn.writeBuf('V')
	b.int32(opType)
	if opType == teeSendKey {
		b.uint32(cekOid)
		b.int32(cekAlgo)
		b.int32(cekLen)
		b.bytes(cek)
	}

	if err := cn.send(b); err != nil {
		return err
	}

	_, _, err := cn.readResponseForEnclave()
	return err
}

func (cn *conn) readResponseForEnclave() (int, []byte, error) {
	var clientKeyLen int
	var clientKey []byte
	for {
		t, r, err := cn.recv1()
		if err != nil {
			cn.setBad()
			return 0, nil, fmt.Errorf("cannot recv from conn: %w", err)
		}
		switch t {
		case 'R':
			rsaKeyLen := r.int32()
			ecdhKeyLen := r.int32()
			rsaKey := string(r.next(rsaKeyLen))
			ecdhKey := string(r.next(ecdhKeyLen))
			if rsaKeyLen == 0 || ecdhKeyLen == 0 || rsaKey == "" || ecdhKey == "" {
				return 0, nil, errors.New("rsa_token or ecdh_token from server is empty!")
			}
			clientKeyLen, clientKey, err = cn.setKeyToClientLogic(rsaKeyLen, ecdhKeyLen, rsaKey, ecdhKey)
			if err != nil {
				return 0, nil, err
			}
		case 'E':
			err = parseError(r, cn)
			{
				var err error
				t, r, err = cn.recv1()
				if err != nil {
					cn.setBad()
					return 0, nil, fmt.Errorf("cannot recv from conn: %w", err)
				}
			}
			fallthrough
		case 'Z':
			cn.processReadyForQuery(r)
			if err != nil {
				return 0, nil, fmt.Errorf("got error from database: %w", err)
			}
			return clientKeyLen, clientKey, nil
		default:
			cn.setBad()
			return 0, nil, fmt.Errorf("unknown response: %q", t)
		}
	}
}

// verifying signatures, save exchange info in PGconn, derived AES shared key, and generate client ecdh key
func (cn *conn) setKeyToClientLogic(rsaKeyLen int, ecdhKeyLen int, rsaKey string, ecdhKey string) (int, []byte, error) {
	conn := cn.pgconn
	cRsaKeyLen := Csize_t(rsaKeyLen)
	cEcdhKeyLen := Csize_t(ecdhKeyLen)
	cRsaKey := CString(rsaKey)
	cEcdhKey := CString(ecdhKey)

	success := set_key_to_client_logic(conn, cRsaKeyLen, cEcdhKeyLen, cRsaKey, cEcdhKey)
	if !success {
		variadic_Cfree(unsafe.Pointer(cRsaKey), unsafe.Pointer(cEcdhKey))

		errstring := GoString(pgconn_errmsg(conn))
		if errstring == "" {
			errstring = "encountered a unreport client error"
		}
		return 0, nil, fmt.Errorf("failed to generate client ecdh token from libpq_ce, %w", errors.New(errstring))
	}

	cClientKey := Cmalloc(Csize_t(1) * Csize_t(unsafe.Sizeof(uintptr(0))))
	cClientKeyLen := Cmalloc(Csize_t(1) * Csize_t(unsafe.Sizeof(uintptr(0))))

	success = get_client_key(conn, (**Cchar)(cClientKey), (*Csize_t)(cClientKeyLen))
	if !success {
		variadic_Cfree(unsafe.Pointer(cRsaKey), unsafe.Pointer(cEcdhKey), cClientKey, cClientKeyLen)

		errstring := GoString(pgconn_errmsg(conn))
		if errstring == "" {
			errstring = "encountered a unreport client error"
		}
		return 0, nil, fmt.Errorf("failed to get client ecdh token from libpq_ce, %w", errors.New(errstring))
	}

	clientKeyLen := int(*(*Csize_t)(cClientKeyLen))
	clientKey := GoBytes(unsafe.Pointer(*(**Cchar)(cClientKey)), Cint(clientKeyLen))

	variadic_Cfree(unsafe.Pointer(cRsaKey), unsafe.Pointer(cEcdhKey), cClientKey, cClientKeyLen)
	return clientKeyLen, clientKey, nil
}

// fetch all ceks from server, decrypt them by cached cmks in client, and re-encrypt them by AES shared key
func (cn *conn) fetchAndDecryptCek() ([]uint, []int, []int, [][]byte, error) {
	conn := cn.pgconn
	cekNum := fetched_columns(conn)

	if cekNum == Cint(0) {
		errstring := GoString(pgconn_errmsg(conn))
		if errstring == "" {
			errstring = "encountered a unreport client error"
		}
		fmt.Println(errors.New(errstring))
		return nil, nil, nil, nil, nil
	}
	if cekNum == Cint(-1) {
		errstring := GoString(pgconn_errmsg(conn))
		if errstring == "" {
			errstring = "encountered a unreport client error"
		}
		return nil, nil, nil, nil, fmt.Errorf("when fetch column ceks, %w", errors.New(errstring))
	}

	cCekOids := Cmalloc(Csize_t(1) * Csize_t(unsafe.Sizeof(uintptr(0))))
	cCekAlgos := Cmalloc(Csize_t(1) * Csize_t(unsafe.Sizeof(uintptr(0))))
	cCekBufLens := Cmalloc(Csize_t(1) * Csize_t(unsafe.Sizeof(uintptr(0))))
	cCekBufs := Cmalloc(Csize_t(1) * Csize_t(unsafe.Sizeof(uintptr(0))))

	success := get_cached_ceks(conn, (**Coid)(cCekOids), (**Cint)(cCekAlgos), (**Cint)(cCekBufLens),
		(***Cchar)(cCekBufs))
	if !success {
		variadic_Cfree(cCekOids, cCekAlgos, cCekBufLens, cCekBufs)

		errstring := GoString(pgconn_errmsg(conn))
		if errstring == "" {
			errstring = "encountered a unreport client error"
		}
		return nil, nil, nil, nil, fmt.Errorf("failed to get cached ceks from pgconn, %w", errors.New(errstring))
	}

	const leftShiftBit = 30
	cCekOidsRef := *(**[uint32(1)<<leftShiftBit - 1]Coid)(cCekOids)
	cCekAlgosRef := *(**[uint32(1)<<leftShiftBit - 1]Cint)(cCekAlgos)
	cCekBufLensRef := *(**[uint32(1)<<leftShiftBit - 1]Cint)(cCekBufLens)
	cCekBufsRef := *(**[uint32(1)<<leftShiftBit - 1]*Cchar)(cCekBufs)

	cekOids := make([]uint, cekNum)
	cekAlgos := make([]int, cekNum)
	cekBufLens := make([]int, cekNum)
	cekBufs := make([][]byte, cekNum)

	var i Cint
	for i = 0; i < cekNum; i++ {
		cekOids[i] = (uint)(cCekOidsRef[i])
		cekAlgos[i] = (int)(cCekAlgosRef[i])
		cekBufLens[i] = (int)(cCekBufLensRef[i])
		cekBufs[i] = GoBytes(unsafe.Pointer(cCekBufsRef[i]), cCekBufLensRef[i])
	}

	variadic_Cfree(cCekOids, cCekAlgos, cCekBufLens, cCekBufs)
	return cekOids, cekAlgos, cekBufLens, cekBufs, nil
}
