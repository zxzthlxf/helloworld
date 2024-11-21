package pq

import (
	"database/sql/driver"
	"fmt"
	"io"
	"reflect"
	"time"

	"gitee.com/opengauss/openGauss-connector-go-pq/oid"
)

const headerSize = 4
const byteaSize = 1073733621
const textSize = 1073741821 // The actual value is less than 1G-1
const byteaColSize = 1073741771

type fieldDesc struct {
	// The object ID of the data type.
	OID oid.Oid
	// The data type size (see pg_type.typlen).
	// Note that negative values denote variable-width types.
	Len int
	// The type modifier (see pg_attribute.atttypmod).
	// The meaning of the modifier is type-specific.
	Mod int
}

func (fd fieldDesc) Type() reflect.Type {
	switch fd.OID {
	case oid.T_int8:
		return reflect.TypeOf(int64(0))
	case oid.T_int4:
		return reflect.TypeOf(int32(0))
	case oid.T_int2:
		return reflect.TypeOf(int16(0))
	case oid.T_varchar, oid.T_text:
		return reflect.TypeOf("")
	case oid.T_bool:
		return reflect.TypeOf(false)
	case oid.T_date, oid.T_time, oid.T_timetz, oid.T_timestamp, oid.T_timestamptz:
		return reflect.TypeOf(time.Time{})
	case oid.T_bytea, oid.T_byteawithoutorderwithequalcol, oid.T_byteawithoutordercol,
		oid.T__byteawithoutorderwithequalcol, oid.T__byteawithoutordercol:
		return reflect.TypeOf([]byte(nil))
	case oid.T_float4:
		return reflect.TypeOf(float32(0))
	case oid.T_float8:
		return reflect.TypeOf(float64(0))
	default:
		return reflect.TypeOf(new(interface{})).Elem()
	}
}

func (fd fieldDesc) Name() string {
	if (fd.OID == oid.T_byteawithoutorderwithequalcol ||
		fd.OID == oid.T_byteawithoutordercol ||
		fd.OID == oid.T__byteawithoutorderwithequalcol ||
		fd.OID == oid.T__byteawithoutordercol) && fd.Mod > -1 {
		return oid.TypeName[oid.Oid(fd.Mod)]
	}
	return oid.TypeName[fd.OID]
}

func (fd fieldDesc) Length() (length int64, ok bool) {
	switch fd.OID {
	case oid.T_text, oid.T_clob:
		return textSize, true
	case oid.T_bytea, oid.T_raw, oid.T_blob:
		return byteaSize, true
	case oid.T_byteawithoutordercol, oid.T_byteawithoutorderwithequalcol, oid.T__byteawithoutordercol, oid.T__byteawithoutorderwithequalcol:
		return byteaColSize, true
	case oid.T_varchar, oid.T_bpchar, oid.T_nvarchar2:
		return int64(fd.Mod - headerSize), true
	default:
		return 0, false
	}
}

func (fd fieldDesc) PrecisionScale() (precision, scale int64, ok bool) {
	switch fd.OID {
	case oid.T_numeric, oid.T__numeric:
		mod := fd.Mod - headerSize
		precision = int64((mod >> 16) & 0xffff)
		scale = int64(mod & 0xffff)
		return precision, scale, true
	default:
		return 0, 0, false
	}
}

type rowsHeader struct {
	colNames []string
	colTyps  []fieldDesc
	colFmts  []format
}

type rows struct {
	cn                      *conn
	finish                  func()
	rowsHeader              //TODO: pointer
	done                    bool
	rb                      readBuf
	result                  driver.Result
	tag                     string
	disable_text_conversion bool

	next *rowsHeader
}

func (rs *rows) Close() error {
	if finish := rs.finish; finish != nil {
		defer finish()
	}
	// no need to look at cn.bad as Next() will
	for {
		err := rs.Next(nil)
		switch err {
		case nil:
		case io.EOF:
			// rs.Next can return io.EOF on both 'Z' (ready for query) and 'T' (row
			// description, used with HasNextResultSet). We need to fetch messages until
			// we hit a 'Z', which is done by waiting for done to be set.
			if rs.done {
				return nil
			}
		default:
			return err
		}
	}
}

func (rs *rows) Columns() []string {
	return rs.colNames
}

func (rs *rows) Result() driver.Result {
	if rs.result == nil {
		return emptyRows
	}
	return rs.result
}

func (rs *rows) Tag() string {
	return rs.tag
}

func (rs *rows) DisableTextConversion() {
	rs.disable_text_conversion = true
}

func (rs *rows) Next(dest []driver.Value) (err error) {
	if rs.done {
		return io.EOF
	}

	cn := rs.cn
	if cn.getBad() {
		return driver.ErrBadConn
	}

	for {
		t, err := cn.recv1Buf(&rs.rb)
		if err != nil {
			cn.setBad()
			return fmt.Errorf("unexpected DataRow after error %s", err)
		}
		switch t {
		case 'E':
			err = parseError(&rs.rb, cn)
		case 'C', 'I':
			if t == 'C' {
				s, err := rs.rb.string()
				if err != nil {
					return fmt.Errorf("cannot get string from read buf: %w", err)
				}
				rs.result, rs.tag, err = cn.parseComplete(s)
				if err != nil {
					return fmt.Errorf("cannot parse complete: %w", err)
				}
			}
			continue
		case 'Z':
			cn.processReadyForQuery(&rs.rb)
			rs.done = true
			if err != nil {
				return err
			}
			return io.EOF
		case 'D':
			n := rs.rb.int16()
			if n < len(dest) {
				dest = dest[:n]
			}
			for i := range dest {
				l := rs.rb.int32()
				if l == -1 {
					dest[i] = nil
					continue
				}

				dest[i], err = decode(&cn.parameterStatus, rs.rb.next(l), rs.colTyps[i].OID, rs.colTyps[i].Mod,
					rs.colFmts[i], rs.disable_text_conversion, rs.rowsHeader.colNames[i], cn.pgconn)
			}
			return err
		case 'T':
			next, err := parsePortalRowDescribe(&rs.rb)
			if err != nil {
				return fmt.Errorf("cannot parse protal row describe: %w", err)
			}
			rs.next = &next
			return io.EOF
		default:
			return fmt.Errorf("unexpected message after execute: %q", t)
		}
	}
}

func (rs *rows) HasNextResultSet() bool {
	hasNext := rs.next != nil && !rs.done
	return hasNext
}

func (rs *rows) NextResultSet() error {
	if rs.next == nil {
		return io.EOF
	}
	rs.rowsHeader = *rs.next
	rs.next = nil
	return nil
}

// ColumnTypeScanType returns the value type that can be used to scan types into.
func (rs *rows) ColumnTypeScanType(index int) reflect.Type {
	return rs.colTyps[index].Type()
}

// ColumnTypeDatabaseTypeName return the database system type name.
func (rs *rows) ColumnTypeDatabaseTypeName(index int) string {
	return rs.colTyps[index].Name()
}

// ColumnTypeLength returns the length of the column type if the column is a
// variable length type. If the column is not a variable length type ok
// should return false.
func (rs *rows) ColumnTypeLength(index int) (length int64, ok bool) {
	return rs.colTyps[index].Length()
}

// ColumnTypePrecisionScale should return the precision and scale for decimal
// types. If not applicable, ok should be false.
func (rs *rows) ColumnTypePrecisionScale(index int) (precision, scale int64, ok bool) {
	return rs.colTyps[index].PrecisionScale()
}
