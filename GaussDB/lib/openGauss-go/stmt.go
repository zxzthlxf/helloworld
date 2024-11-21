/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2024. All rights reserved.
 * Code lines related to stmt
 */

package pq

import (
	"database/sql/driver"
	"fmt"
	"gitee.com/opengauss/openGauss-connector-go-pq/oid"
)

type stmt struct {
	cn   *conn
	name string
	rowsHeader
	colFmtData []byte
	paramTypes []oid.Oid
	closed     bool
}

func (st *stmt) Close() (err error) {
	if st.closed {
		return nil
	}
	if st.cn.getBad() {
		return driver.ErrBadConn
	}

	w := st.cn.writeBuf('C')
	w.byte('S')
	w.string(st.name)
	if err = st.cn.send(w); err != nil {
		return fmt.Errorf("fail to send: %w", err)
	}

	if err = st.cn.send(st.cn.writeBuf('S')); err != nil {
		return fmt.Errorf("fail to send: %w", err)
	}

	t, _, err := st.cn.recv1()
	if err != nil {
		st.cn.setBad()
		return fmt.Errorf("cannot recv from conn: %w", err)
	}
	if t != '3' {
		st.cn.setBad()
		return fmt.Errorf("unexpected close response: %q", t)
	}
	st.closed = true

	t, r, err := st.cn.recv1()
	if err != nil {
		st.cn.setBad()
		return fmt.Errorf("cannot recv from conn: %w", err)
	}
	if t != 'Z' {
		st.cn.setBad()
		return fmt.Errorf("expected ready for query, but got: %q", t)
	}
	st.cn.processReadyForQuery(r)

	return nil
}

func (st *stmt) Query(v []driver.Value) (driver.Rows, error) {
	return st.query(v)
}

func (st *stmt) query(v []driver.Value) (r *rows, err error) {
	if st.cn.getBad() {
		return nil, driver.ErrBadConn
	}

	err = st.exec(v, true)
	if err != nil {
		return nil, fmt.Errorf("cannot exec with value %v: %v", v, err)
	}

	return &rows{
		cn:         st.cn,
		rowsHeader: st.rowsHeader,
	}, nil
}

const maxParamNum = 65536 // maximum number of parameters bound in a batch

func (st *stmt) Exec(v []driver.Value) (driver.Result, error) {
	var err error

	if st.cn.getBad() {
		return nil, driver.ErrBadConn
	}

	if err = st.exec(v, true); err != nil {
		return nil, fmt.Errorf("fail to exec, error: %w", err)
	}
	res, _, err := st.cn.readExecuteResponse("simple query")
	if err != nil {
		return nil, fmt.Errorf("cannot read execute response: %w", err)
	}

	return res, nil
}
