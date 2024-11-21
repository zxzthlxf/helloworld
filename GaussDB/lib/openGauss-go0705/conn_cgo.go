package pq

import (
	"database/sql/driver"
	"errors"
	"reflect"
	"strconv"
	"unsafe"
)

// getResultSet - returns the entire resultset in string representation
//  @param rows - rows object to read the data from the database server
//  @param columns_count - number of columns in the result set
//  @return resultset
//  @return err
func getResultSet(rows *rows, columns_count int) (resultset [][]string, err error) {

	// prepare row buffer to pass to rows.Next() to fill
	driver_value_row_strings := make([]string, columns_count, columns_count)
	driver_value_row := make([]driver.Value, cap(driver_value_row_strings))
	for i, s := range driver_value_row_strings {
		driver_value_row[i] = s
	}

	rows.DisableTextConversion()
	for err := rows.Next(driver_value_row); rows.done == false; err = rows.Next(driver_value_row) {
		if err != nil {
			return nil, errors.New("failed to retrieve next row")
		}

		row_strings := make([]string, columns_count)
		for i, field := range driver_value_row {
			var dest string
			switch s := field.(type) {
			case string:
				dest = s
			case []byte:
				dest = string(s)
			default:
				rv := reflect.ValueOf(s)
				switch rv.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					dest = strconv.FormatInt(rv.Int(), 10)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					dest = strconv.FormatUint(rv.Uint(), 10)
				case reflect.Float64:
					dest = strconv.FormatFloat(rv.Float(), 'g', -1, 64)
				case reflect.Float32:
					dest = strconv.FormatFloat(rv.Float(), 'g', -1, 32)
				case reflect.Bool:
					dest = strconv.FormatBool(rv.Bool())
				default:
					dest = ""
				}
			}
			if i < columns_count {
				row_strings[i] = dest
			}
		}
		resultset = append(resultset, row_strings)
	}
	return resultset, nil
}

// convertResultSetToC - convert a resultset in golang to C
//  @param resultset - resultset in string representation
//  @param rows - rows object to read the data from the database server
//  @param columns_count - number of columns in the result set
//  @return ***Cchar 
func convertResultSetToC(resultset [][]string, rows_count int, columns_count int) ***Cchar {
	c_resultset := Cmalloc(Csize_t(rows_count) * Csize_t(unsafe.Sizeof(uintptr(0))))
	c_resultset_ref := (*[1<<30 - 1]*Cchar)(c_resultset)
	for i, row := range resultset {
		c_row := Cmalloc(Csize_t(columns_count) * Csize_t(unsafe.Sizeof(uintptr(0))))
		c_resultset_ref[i] = (*Cchar)(unsafe.Pointer(c_row))
		c_row_ref := (*[1<<30 - 1]*Cchar)(c_row)
		for j, field := range row {
			c_row_ref[j] = CString(field)
		}
	}
	return (***Cchar)(unsafe.Pointer(c_resultset))
}

// convertColumnsToC - convert columns list to C structure
//  @param colNames 
//  @param columns_count - number of columns in the result set
//  @return **Cchar 
func convertColumnsToC(colNames []string) **Cchar {
	c_column_names := Cmalloc(Csize_t(len(colNames)) * Csize_t(unsafe.Sizeof(uintptr(0))))
        c_column_names_ref := (*[1<<30 - 1]*Cchar)(c_column_names)
	for i, colname := range colNames {
		c_column_names_ref[i] = CString(colname)
	}
	
	return (**Cchar)(unsafe.Pointer(c_column_names))
}

//export Conncgo_SimpleQuery
func Conncgo_SimpleQuery_func(conn_ptr unsafe.Pointer, query string) (result ***Cchar, 
	columns_names **Cchar, rows_count int, columns_count int) {
	conn := (getPointer(conn_ptr).(*conn))
	if conn == nil {
		return nil, nil, 0, 0
	}
	rows, err := conn.simpleQuery(query)
	if err != nil || rows == nil {
		return nil, nil, 0, 0
	}
	columns_count = len(rows.rowsHeader.colNames)
	resultset, err := getResultSet(rows, columns_count)
	if err != nil {
		return nil, nil, 0, 0
	}
	rows_count = len(resultset)

	c_column_names := convertColumnsToC(rows.rowsHeader.colNames)
	c_resultset := convertResultSetToC(resultset, rows_count, columns_count)
	return c_resultset, c_column_names, rows_count, columns_count
}
