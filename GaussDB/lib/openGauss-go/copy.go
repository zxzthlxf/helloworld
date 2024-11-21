package pq

import (
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

var (
	errCopyInClosed               = errors.New("pq: copyin statement has already been closed")
	errBinaryCopyNotSupported     = errors.New("pq: only text format supported for COPY")
	errCopyToNotSupported         = errors.New("pq: COPY TO is not supported")
	errCopyNotSupportedOutsideTxn = errors.New("pq: COPY is only allowed inside a transaction")
	errCopyInProgress             = errors.New("pq: COPY in progress")
)

// CopyIn creates a COPY FROM statement which can be prepared with
// Tx.Prepare().  The target table should be visible in search_path.
func CopyIn(table string, columns ...string) string {
	stmt := "COPY " + QuoteIdentifier(table) + " ("
	for i, col := range columns {
		if i != 0 {
			stmt += ", "
		}
		stmt += QuoteIdentifier(col)
	}
	stmt += ") FROM STDIN"
	return stmt
}

// CopyInSchema creates a COPY FROM statement which can be prepared with
// Tx.Prepare().
func CopyInSchema(schema, table string, columns ...string) string {
	stmt := "COPY " + QuoteIdentifier(schema) + "." + QuoteIdentifier(table) + " ("
	for i, col := range columns {
		if i != 0 {
			stmt += ", "
		}
		stmt += QuoteIdentifier(col)
	}
	stmt += ") FROM STDIN"
	return stmt
}

type copyin struct {
	cn      *conn
	buffer  []byte
	rowData chan []byte
	done    chan bool
	driver.Result

	closed bool

	sync.Mutex // guards err
	err        error
}

const ciBufferSize = 64 * 1024

// flush buffer before the buffer is filled up and needs reallocation
const ciBufferFlushSize = 63 * 1024

func (cn *conn) prepareCopyIn(q string) (_ driver.Stmt, err error) { // TODO: named return value
	if !cn.isInTransaction() {
		return nil, errCopyNotSupportedOutsideTxn
	}

	ci := &copyin{
		cn:      cn,
		buffer:  make([]byte, 0, ciBufferSize),
		rowData: make(chan []byte),
		done:    make(chan bool, 1),
	}
	// add CopyData identifier + 4 bytes for message length
	ci.buffer = append(ci.buffer, 'd', 0, 0, 0, 0)

	if cn.pgconn != nil {
		var queryCstring *Cchar
		runtime.LockOSThread()
		q, queryCstring, err = cn.replaceQuery("", q)
		defer Cfree(unsafe.Pointer(queryCstring))
		if err != nil {
			runtime.UnlockOSThread()
			return nil, fmt.Errorf("cannot replace query: %w", err)
		}
		/*
			The copy protocol works in the same way as unnamed prepared statement.
			the backend starts responding right away with COPY messages and only sends a RFQ in the end.
		*/
		accept_pending_statements(ci.cn.pgconn)
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

awaitCopyInResponse:
	for {
		t, r, err := cn.recv1()
		if err != nil {
			cn.setBad()
			return nil, fmt.Errorf("cannot recv from conn: %w", err)
		}
		switch t {
		case 'G': // CopyInResponse
			if r.byte() != 0 {
				err = errBinaryCopyNotSupported
				break awaitCopyInResponse
			}
			go ci.resploop()
			return ci, nil
		case 'H': // CopyOutResponse
			err = errCopyToNotSupported
			break awaitCopyInResponse
		case 'E': // ErrorResponse
			err = parseError(r, cn)
		case 'Z': // ReadyForQuery
			if err == nil {
				ci.setBad()
				return nil, fmt.Errorf("unexpected ReadyForQuery in response to COPY")
			}
			cn.processReadyForQuery(r)
			return nil, err
		default:
			ci.setBad()
			return nil, fmt.Errorf("unknown response for copy query: %q", t)
		}
	}

	// something went wrong, abort COPY before we return
	b = cn.writeBuf('f')
	b.string(err.Error())
	if err = cn.send(b); err != nil {
		return nil, fmt.Errorf("fail to send: %w", err)
	}

	for {
		t, r, err := cn.recv1()
		if err != nil {
			cn.setBad()
			return nil, fmt.Errorf("cannot recv from conn: %w", err)
		}
		switch t {
		case 'c', 'C', 'E':
		case 'Z':
			// correctly aborted, we're done
			cn.processReadyForQuery(r)
			return nil, err
		default:
			ci.setBad()
			return nil, fmt.Errorf("unknown response for CopyFail: %q", t)
		}
	}
}

func (ci *copyin) flush(buf []byte) error {
	if ci.cn.pgconn != nil {
		buf_adv := buf[5:]
		c_buf_adv := CBytes(buf_adv)
		defer Cfree(unsafe.Pointer(c_buf_adv))
		buf_adv_len := len(buf_adv)
		var out_buffer *Cchar
		ret := process_copy_chunk(ci.cn.pgconn, (*Cchar)(c_buf_adv),
			(Cint)(buf_adv_len), (**Cchar)(&out_buffer))
		defer Cfree(unsafe.Pointer(out_buffer))

		if ret > 0 {
			buf = buf[0:5]
			append_data := GoBytes(unsafe.Pointer(out_buffer), ret)
			buf = append(buf, append_data...)
		} else if ret < 0 {
			return nil
		}
	}
	// set message length (without message identifier)
	binary.BigEndian.PutUint32(buf[1:], uint32(len(buf)-1))

	if _, err := ci.cn.c.Write(buf); err != nil {
		return connErr{
			msg: fmt.Sprintf("fail to write: %v", err),
			err: driver.ErrBadConn,
		}
	}

	return nil
}

func (ci *copyin) resploop() error {
	for {
		var r readBuf
		t, err := ci.cn.recvMessage(&r)
		if err != nil {
			ci.setBad()
			ci.setError(err)
			ci.done <- true
			return nil
		}
		switch t {
		case 'C':
			// complete
			s, err := r.string()
			if err != nil {
				return fmt.Errorf("cannot get string from read buf: %w", err)
			}
			res, _, err := ci.cn.parseComplete(s)
			if err != nil {
				return fmt.Errorf("cannot parse complete: %w", err)
			}
			ci.setResult(res)
		case 'N':
			if n := ci.cn.noticeHandler; n != nil {
				n(parseError(&r, ci.cn))
			}
		case 'Z':
			ci.cn.processReadyForQuery(&r)
			ci.done <- true
			return nil
		case 'E':
			err := parseError(&r, ci.cn)
			ci.setError(err)
		default:
			ci.setBad()
			ci.setError(fmt.Errorf("unknown response during CopyIn: %q", t))
			ci.done <- true
			return nil
		}
	}
}

func (ci *copyin) setBad() {
	ci.Lock()
	ci.cn.setBad()
	ci.Unlock()
}

func (ci *copyin) isBad() bool {
	ci.Lock()
	b := ci.cn.getBad()
	ci.Unlock()
	return b
}

func (ci *copyin) isErrorSet() bool {
	ci.Lock()
	isSet := (ci.err != nil)
	ci.Unlock()
	return isSet
}

// setError() sets ci.err if one has not been set already.  Caller must not be
// holding ci.Mutex.
func (ci *copyin) setError(err error) {
	ci.Lock()
	if ci.err == nil {
		ci.err = err
	}
	ci.Unlock()
}

func (ci *copyin) setResult(result driver.Result) {
	ci.Lock()
	ci.Result = result
	ci.Unlock()
}

func (ci *copyin) getResult() driver.Result {
	ci.Lock()
	result := ci.Result
	ci.Unlock()
	if result == nil {
		return driver.RowsAffected(0)
	}
	return result
}

func (ci *copyin) NumInput() int {
	return -1
}

func (ci *copyin) Query(v []driver.Value) (r driver.Rows, err error) {
	return nil, ErrNotSupported
}

// Exec inserts values into the COPY stream. The insert is asynchronous
// and Exec can return errors from previous Exec calls to the same
// COPY stmt.
//
// You need to call Exec(nil) to sync the COPY stream and to get any
// errors from pending data, since Stmt.Close() doesn't return errors
// to the user.
func (ci *copyin) Exec(v []driver.Value) (r driver.Result, err error) {
	if ci.closed {
		return nil, errCopyInClosed
	}

	if ci.isBad() {
		return nil, driver.ErrBadConn
	}

	if ci.isErrorSet() {
		return nil, ci.err
	}

	if len(v) == 0 {
		if err := ci.Close(); err != nil {
			return driver.RowsAffected(0), err
		}

		return ci.getResult(), fmt.Errorf("\"COPY FROM STDIN\" requires the given parameters")
	}

	numValues := len(v)
	for i, value := range v {
		ci.buffer, err = appendEncodedText(&ci.cn.parameterStatus, ci.buffer, value)
		if err != nil {
			return nil, fmt.Errorf("cannot append encoded test: %w", err)
		}
		if i < numValues-1 {
			ci.buffer = append(ci.buffer, '\t')
		}
	}

	ci.buffer = append(ci.buffer, '\n')

	if len(ci.buffer) > ciBufferFlushSize {
		if err = ci.flush(ci.buffer); err != nil {
			return nil, fmt.Errorf("fail to flash: %w", err)
		}
		// reset buffer, keep bytes for message identifier and length
		ci.buffer = ci.buffer[:5]
	}

	return driver.RowsAffected(0), nil
}

func (ci *copyin) Close() (err error) {
	if ci.closed { // Don't do anything, we're already closed
		return nil
	}
	ci.closed = true

	if ci.isBad() {
		return driver.ErrBadConn
	}

	if len(ci.buffer) > 0 {
		if err = ci.flush(ci.buffer); err != nil {
			return fmt.Errorf("fail to flash: %w", err)
		}
	}
	// Avoid touching the scratch buffer as resploop could be using it.
	if err = ci.cn.sendSimpleMessage('c'); err != nil {
		return fmt.Errorf("cannot send simple message: %w", err)
	}

	<-ci.done
	ci.cn.inCopy = false

	if ci.isErrorSet() {
		return ci.err
	}
	return nil
}
