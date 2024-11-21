package pq

import (
	"bufio"
	"context"
	"crypto/tls"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	targetSessionAttrsAny uint8 = iota
	targetSessionAttrsReadWrite
	targetSessionAttrsReadOnly
	targetSessionAttrsMaster
	targetSessionAttrsSlave
	targetSessionAttrsPreferSlave
)

// Compile time validation that our types implement the expected interfaces
var (
	_ driver.Driver = Driver{}
)

// Driver is the Postgres database driver.
type Driver struct{}

func init() {
	goVer := runtime.Version()
	var v1, v2, v3 int
	var err error
	_, err = fmt.Sscanf(goVer, "go%d.%d.%d", &v1, &v2, &v3)
	if err != nil {
		_, err = fmt.Sscanf(goVer, "go%d.%d", &v1, &v2)
	}
	if err == nil {
		if v1 < 1 || v2 < 13 {
			log.Println("Your go version is earlier, please upgrade to at least go1.13")
		}
	}
	sql.Register("opengauss", &Driver{})
	sql.Register("postgres", &Driver{})
	sql.Register("postgresql", &Driver{})
	sql.Register("mogdb", &Driver{})
}

func (d Driver) OpenConnector(dsn string) (driver.Connector, error) {
	return NewConnector(dsn)
}

// Open opens a new connection to the database. name is a connection string.
// Most users should only use it through database/sql package from the standard
// library.
func (d Driver) Open(name string) (driver.Conn, error) {
	return Open(name)
}

// DialFunc is a function that can be used to connect to a PostgreSQL server.
type DialFunc func(ctx context.Context, network, addr string) (net.Conn, error)

// BuildFrontendFunc is a function that can be used to create Frontend implementation for connection.
// type BuildFrontendFunc func(r io.Reader, w io.Writer) Frontend

// LookupFunc is a function that can be used to lookup IPs addrs from host.
type LookupFunc func(ctx context.Context, host string) (addrs []string, err error)

// Dialer is the dialer interface. It can be used to obtain more control over
// how pq creates network connections.
type Dialer interface {
	Dial(network, address string) (net.Conn, error)
	DialTimeout(network, address string, timeout time.Duration) (net.Conn, error)
}

// DialerContext is the context-aware dialer interface.
type DialerContext interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type defaultDialer struct {
	d net.Dialer
}

func (d defaultDialer) Dial(network, address string) (net.Conn, error) {
	return d.d.Dial(network, address)
}
func (d defaultDialer) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return d.DialContext(ctx, network, address)
}
func (d defaultDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.d.DialContext(ctx, network, address)
}

// Connector represents a fixed configuration for the pq driver with a given
// name. Connector satisfies the database/sql/driver Connector interface and
// can be used to create any number of DB Conn's via the database/sql OpenDB
// function.
//
// See https://golang.org/pkg/database/sql/driver/#Connector.
// See https://golang.org/pkg/database/sql/#OpenDB.
type Connector struct {
	dialer connectorDialer
	config *Config
}

// Open opens a new connection to the database. dsn is a connection string.
// Most users should only use it through database/sql package from the standard
// library.
func Open(dsn string) (_ driver.Conn, err error) {
	return DialOpen(dsn)
}

// DialOpen opens a new connection to the database using a dialer.
func DialOpen(dsn string) (_ driver.Conn, err error) {
	c, err := NewConnector(dsn)
	if err != nil {
		return nil, err
	}
	return c.open(context.Background())
}

// NewConnector returns a connector for the pq driver in a fixed configuration
// with the given dsn. The returned connector can be used to create any number
// of equivalent Conn's. The returned connector is intended to be used with
// database/sql.OpenDB.
//
// See https://golang.org/pkg/database/sql/driver/#Connector.
// See https://golang.org/pkg/database/sql/#OpenDB.
func NewConnector(dsn string) (*Connector, error) {
	cfg, distCfg, err := ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	var balancer cnsBalancer
	balPol := distCfg.balancePolicy
	cn := &Connector{config: cfg}

	if balPol == balanceNone { // single 模式
		cn.dialer = &singleDialer{
			dialer: defaultDialer{},
		}
		return cn, nil
	}

	// load balancer from balance policy
	switch balPol {
	case balanceRoundRobin, balanceLeastConn:
		balancer = &roundRobbinBalancer{
			startIdx:        1,
			idxLock:         &sync.Mutex{},
			shuffleBalancer: &shuffleBalancer{},
		}
	case balancePriority:
		balancer = &priorityBalancer{
			num: distCfg.priorityNum,
			roundRobbinBalancer: &roundRobbinBalancer{
				startIdx:        1,
				idxLock:         &sync.Mutex{},
				shuffleBalancer: &shuffleBalancer{},
			},
			urlCNs: deepCopyCnsFromConfig(cfg),
		}
	case balanceShuffle:
		balancer = &shuffleBalancer{}
	default:
	}

	// support distribute
	cn.dialer = &distributeDialer{
		dialer:                defaultDialer{},
		cnsBalancer:           balancer,
		refreshCNsIntervalSec: distCfg.refreshCNsIntervalSec,
		cnsLock:               &sync.RWMutex{},
		usingEip:              distCfg.isUsingEip,
		logger:                cfg.Logger,
		logLevel:              cfg.LogLevel,
		tlsCfgs:               distCfg.tlsCfgs,
	}

	db := sql.OpenDB(cn) // TODO: move into refreshCNs method

	if distDia, ok := cn.dialer.(*distributeDialer); ok {
		if err = distDia.doRefreshCNs(context.Background(), db); err != nil {
			return nil, fmt.Errorf("cannot refresh cns: %w", err)
		}
		go distDia.refreshCNs(context.Background(), db)
	}

	return cn, nil
}

func deepCopyCnsFromConfig(cfg *Config) []coordinateNode {
	var tmpCNodes []coordinateNode
	tmpCNodes = append(tmpCNodes, coordinateNode{
		ip:   cfg.Host,
		port: cfg.Port,
	})
	for _, fallback := range cfg.Fallbacks {
		tmpCNodes = append(tmpCNodes, coordinateNode{
			ip:   fallback.Host,
			port: fallback.Port,
		})
	}
	return removeDuplicateNodes(tmpCNodes)
}

// Connect returns a connection to the database using the fixed configuration
// of this Connector. Context is not used.
func (c *Connector) Connect(ctx context.Context) (driver.Conn, error) {
	return c.open(ctx)
}

// Driver returns the underlying driver of this Connector.
func (c *Connector) Driver() driver.Driver {
	return &Driver{}
}

func (c *Connector) open(ctx context.Context) (cn *conn, err error) {
	if !c.config.createdByParseConfig {
		return nil, errors.New("config must be created by ParseConfig")
	}
	return c.dialer.dial(ctx, c.config)
}

type connectorDialer interface {
	dial(ctx context.Context, config *Config) (cn *conn, err error)
}

type singleDialer struct {
	dialer Dialer
}

func (s *singleDialer) dial(ctx context.Context, config *Config) (cn *conn, err error) {
	// ConnectTimeout restricts the whole connection process.
	//defer errRecoverNoErrBadConn(&err)
	if config.ConnectTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.ConnectTimeout)
		defer cancel()
	}
	// Simplify usage by treating primary config and fallbacks the same.
	fallbackConfigs := []*FallbackConfig{
		{
			Host:      config.Host,
			Port:      config.Port,
			TLSConfig: config.TLSConfig,
		},
	}
	fallbackConfigs = append(fallbackConfigs, config.Fallbacks...)

	fallbackConfigs, err = expandWithIPs(ctx, config.LookupFunc, fallbackConfigs)
	if err != nil {
		return nil, &connectError{config: config, msg: "hostname resolving error", err: err}
	}

	if len(fallbackConfigs) == 0 {
		return nil, &connectError{config: config, msg: "hostname resolving error",
			err: errors.New("ip addr wasn't found")}
	}
	var masterConn *conn = nil
	for _, fc := range fallbackConfigs {
		cn, err = connectFallbackConfig(ctx, config, fc)
		if err != nil {
			if pgErr, ok := err.(*Error); ok {
				err = &connectError{config: config, msg: "server error", err: pgErr}
				ErrCodeInvalidPassword := "28P01"                   // worng password
				ErrCodeInvalidAuthorizationSpecification := "28000" // db does not exist
				if pgErr.Code.String() == ErrCodeInvalidPassword ||
					pgErr.Code.String() == ErrCodeInvalidAuthorizationSpecification {
					break
				}
			}
			config.Log(context.Background(), LogLevelDebug, fmt.Sprintf(
				"%s, %v:%v",
				err.Error(),
				fc.Host,
				fc.Port),
				map[string]interface{}{})
			cn = nil
			continue
		}
		if cn.isMasterForPreferSlave {
			if masterConn == nil {
				masterConn = cn
			} else {
				err := cn.Close()
				if err != nil {
					return nil, fmt.Errorf("cannot close connector: %v", err)
				}
			}
			cn = nil
			continue
		}
		config.Log(context.Background(), LogLevelDebug,
			fmt.Sprintf("find instance: (%v:%v)", fc.Host, fc.Port),
			map[string]interface{}{})
		break
	}

	if cn == nil {
		if masterConn != nil {
			config.Log(context.Background(), LogLevelDebug, "using master when perferSlave", map[string]interface{}{})
			masterConn.disablePreparedBinaryResult = config.disablePreparedBinaryResult
			masterConn.binaryParameters = config.binaryParameters
			return masterConn, nil
		}

		return nil, fmt.Errorf("connect failed. please check connect string, err:%s", err.Error())
	}
	if masterConn != nil {
		err := masterConn.Close()
		if err != nil {
			return nil, fmt.Errorf("cannot close master connect: %v", err)
		}
	}
	if err != nil {
		return nil, err // no need to wrap in connectError because it will already be wrapped in all cases except PgError
	}
	cn.disablePreparedBinaryResult = config.disablePreparedBinaryResult
	cn.binaryParameters = config.binaryParameters
	return cn, nil
}

func connectFallbackConfig(ctx context.Context, config *Config, fallbackConfig *FallbackConfig) (cn *conn, err error) {
	cn = &conn{
		config:         config,
		logLevel:       config.LogLevel,
		logger:         config.Logger,
		fallbackConfig: fallbackConfig,
	}
	cn.log(ctx, LogLevelInfo, fmt.Sprintf(
		"Dialing server: (%v:%v)",
		fallbackConfig.Host,
		fallbackConfig.Port),
		map[string]interface{}{})
	network, address := NetworkAddress(fallbackConfig.Host, fallbackConfig.Port)
	cn.c, err = config.DialFunc(ctx, network, address) // exact establish net connection
	if err != nil {
		return nil, &connectError{config: config, msg: "dial error", err: err}
	}
	if fallbackConfig.TLSConfig != nil {
		if err := cn.startTLS(fallbackConfig.TLSConfig); err != nil {
			if err := cn.c.Close(); err != nil {
				return nil, &connectError{config: config, msg: "close connect error", err: err}
			}
			return nil, &connectError{config: config, msg: "tls error", err: err}
		}
	}

	cn.buf = bufio.NewReader(cn.c)
	if err = cn.startup(); err != nil {
		_ = cn.Close()
		return nil, fmt.Errorf("fail to startup: %w", err)
	}

	// reset the deadline, in case one was set (see dial)
	if config.ConnectTimeout.Seconds() > 0 {
		if err = cn.c.SetDeadline(time.Time{}); err != nil {
			_ = cn.Close()
			return nil, fmt.Errorf("cannot set deadline: %w", err)
		}
	}
	return cn, err
}

type validateError string

func (v validateError) Error() string {
	return string(v)
}

func deepCopyCnsInPrimaryCluster(ctx context.Context, cfg *Config, usingEip bool) ([]coordinateNode, error) {
	urlCNs := deepCopyCnsFromConfig(cfg)
	if len(urlCNs) == 1 {
		return urlCNs, nil
	}

	var query string
	if usingEip {
		query = "select node_host1, node_port1 from pgxc_node where node_type='C' and nodeis_active = true order by node_host1;"
	} else {
		query = "select node_host,node_port from pgxc_node where node_type='C' and nodeis_active = true order by node_host;"
	}
	cn := &Connector{
		dialer: &singleDialer{dialer: defaultDialer{}},
		config: cfg,
	}
	db := sql.OpenDB(cn)
	defer db.Close()
	res, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("fail to query: %w", err)
	}

	var cns []coordinateNode
	rowNum := 1
	for res.Next() {
		var host string
		var port uint16
		err = res.Scan(&host, &port)
		if err != nil {
			return nil, fmt.Errorf("failed to parse the query result: %v", err)
		}
		cns = append(cns, coordinateNode{
			ip:   host,
			port: port,
		})
		cfg.Log(ctx, LogLevelDebug, fmt.Sprintf(
			"received CN Info: %v - %v:%v",
			rowNum,
			host,
			port),
			map[string]interface{}{})
		rowNum++
	}
	if len(cns) == 0 {
		return nil, validateError("want coordinate nodes; got data node")
	}

	// filter current cluster from urlCNs
	m := make(map[string]struct{})
	for _, cn := range cns {
		if cn.ip == "localhost" {
			cn.ip = "127.0.0.1"
		}
		m[fmt.Sprintf("%s:%d", cn.ip, cn.port)] = struct{}{}
	}
	for i := 0; i < len(urlCNs); {
		if _, ok := m[fmt.Sprintf("%s:%d", urlCNs[i].ip, urlCNs[i].port)]; !ok {
			copy(urlCNs[i:], urlCNs[i+1:])
			urlCNs = urlCNs[:len(urlCNs)-1]
		} else {
			i++
		}
	}

	return urlCNs, nil
}

func expandWithIPs(ctx context.Context, lookupFn LookupFunc, fallbacks []*FallbackConfig) ([]*FallbackConfig, error) {
	var configs []*FallbackConfig

	for _, fb := range fallbacks {
		// skip resolve for unix sockets
		if strings.HasPrefix(fb.Host, "/") {
			configs = append(configs, &FallbackConfig{
				Host:      fb.Host,
				Port:      fb.Port,
				TLSConfig: fb.TLSConfig,
			})

			continue
		}

		ips, err := lookupFn(ctx, fb.Host)
		if err != nil {
			return nil, err
		}

		for _, ip := range ips {
			configs = append(configs, &FallbackConfig{
				Host:      ip,
				Port:      fb.Port,
				TLSConfig: fb.TLSConfig,
			})
		}
	}

	return configs, nil
}

type distributeDialer struct {
	refreshCNsIntervalSec int
	usingEip              bool
	cnsLock               *sync.RWMutex
	coordinateNodes       []coordinateNode
	cnsBalancer           cnsBalancer

	tlsCfgs []*tls.Config
	dialer  Dialer

	logger   Logger
	logLevel LogLevel
}

func (d *distributeDialer) Log(ctx context.Context, level LogLevel, msg string, data map[string]interface{}) {
	if d.logger != nil && d.logLevel >= level {
		d.logger.Log(ctx, level, msg, data)
	}
}

func (d *distributeDialer) dial(ctx context.Context, cfg *Config) (*conn, error) {
	cns := make([]coordinateNode, len(d.coordinateNodes))
	d.cnsLock.RLock()
	copy(cns, d.coordinateNodes)
	d.cnsLock.RUnlock()

	// TODO: optimise
	// For the first connection, the original connection in config is used, which is not applicable to cns.
	// The original config can be used to deduplicate and encapsulate cns,
	if len(cns) == 0 {
		var tmpCNodes []coordinateNode
		tmpCNodes = append(tmpCNodes, coordinateNode{
			ip:   cfg.Host,
			port: cfg.Port,
		})
		for _, config := range cfg.Fallbacks {
			tmpCNodes = append(tmpCNodes, coordinateNode{
				ip:   config.Host,
				port: config.Port,
			})
		}
		cns = removeDuplicateNodes(tmpCNodes)

		if bal, ok := d.cnsBalancer.(*priorityBalancer); ok {
			bal.urlCNs = cns
		}
	}
	if balancer := d.cnsBalancer.balance; balancer != nil {
		roundIdx := d.cnsBalancer.balance(cns)
		if len(cns) > 0 && d.logLevel >= LogLevelDebug {
			var build strings.Builder
			build.WriteString(fmt.Sprintf("after balance, roundIdx: %v, seq: [", roundIdx))
			for _, node := range cns {
				build.WriteString(fmt.Sprintf(" %v:%v,", node.ip, node.port))
			}
			build.WriteString("]")
			d.Log(ctx, LogLevelDebug, build.String(), map[string]interface{}{})
		}
	}
	sslMode := os.Getenv("PGSSLMODE")
	tlsCfgs := d.tlsCfgs
	var coorNodes []coordinateNode
	for _, cNode := range cns {
		for _, tlsConfig := range tlsCfgs {
			if sslMode == "verify-full" {
				tlsConfig.ServerName = cNode.ip
			}
			coorNodes = append(coorNodes, coordinateNode{
				ip:        cNode.ip,
				port:      cNode.port,
				TLSConfig: tlsConfig,
			})
		}
	}
	if cfg.ConnectTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.ConnectTimeout)
		defer cancel()
	}
	// Simplify usage by treating primary cfg and fallbacks the same.\
	coorNodes, err := expandWithIPsDist(ctx, cfg.LookupFunc, coorNodes)
	if err != nil {
		return nil, &connectError{msg: "hostname resolving error", err: err}
	}
	if len(coorNodes) == 0 {
		return nil, &connectError{msg: "hostname resolving error",
			err: errors.New("ip addr wasn't found")}
	}
	cn := &conn{}
	for _, cNode := range coorNodes { // TODO: fetch with singleDialer.dial
		cfg.Host = cNode.ip
		cfg.Port = cNode.port
		cn, err = connectCNodeConfig(ctx, cfg, cNode) // TODO: refactor error handling
		if err != nil {
			if pgErr, ok := err.(*Error); ok {
				err = &connectError{config: cfg, msg: "server error", err: pgErr}
				ErrCodeInvalidPassword := "28P01"                   // worng password
				ErrCodeInvalidAuthorizationSpecification := "28000" // db does not exist
				if pgErr.Code.String() == ErrCodeInvalidPassword ||
					pgErr.Code.String() == ErrCodeInvalidAuthorizationSpecification {
					break
				}
			}
			cfg.Log(context.Background(), LogLevelInfo, fmt.Sprintf(
				"fail to dial: %v, host: (%v:%v)",
				err,
				cNode.ip,
				cNode.port),
				map[string]interface{}{})
			continue
		}
		cfg.Log(context.Background(), LogLevelDebug,
			fmt.Sprintf("find instance: (%v:%v)", cNode.ip, cNode.port),
			map[string]interface{}{})
		break
	}
	if err != nil {
		return nil, err // no need to wrap in connectError because it will already be wrapped in all cases except PgError
	}
	cn.disablePreparedBinaryResult = cfg.disablePreparedBinaryResult
	cn.binaryParameters = cfg.binaryParameters
	return cn, nil
}

func removeDuplicateNodes(a []coordinateNode) (ret []coordinateNode) { // todo: optimise
	n := len(a)
	for i := 0; i < n; i++ {
		state := false
		for j := i + 1; j < n; j++ {
			if j > 0 && reflect.DeepEqual(a[i], a[j]) {
				state = true
				break
			}
		}
		if !state {
			ret = append(ret, a[i])
		}
	}
	return ret
}

func (d *distributeDialer) refreshCNs(ctx context.Context, db *sql.DB) {
	refreshTime := 10 * time.Second
	if d.refreshCNsIntervalSec > 0 {
		refreshTime = time.Duration(d.refreshCNsIntervalSec) * time.Second
	}
	d.Log(ctx, LogLevelInfo,
		fmt.Sprintf("Start the goroutine of refreshing CN list. refreshTime : %v", refreshTime),
		map[string]interface{}{})
	defer d.Log(ctx, LogLevelInfo, "End the goroutine of refreshing CN list.", map[string]interface{}{})

	t := time.NewTicker(refreshTime)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			err := db.Close()
			if err != nil {
				d.Log(ctx, LogLevelError, fmt.Sprintf("cannot close database: %v", err), nil)
			}
			return
		case _, _ = <-t.C:
			err := d.doRefreshCNs(ctx, db)
			if err != nil {
				d.Log(ctx, LogLevelError,
					fmt.Sprintf("Failed to query the CN list, error info: %v", err.Error()), map[string]interface{}{})
				if err.Error() == "sql: database is closed" {
					return
				}
			}
		}
	}
}

func (d *distributeDialer) doRefreshCNs(ctx context.Context, db *sql.DB) error {
	var queryStr string
	if d.usingEip {
		queryStr = "select node_host1, node_port1 from pgxc_node where node_type='C' and nodeis_active = true order by node_host1;"
	} else {
		queryStr = "select node_host,node_port from pgxc_node where node_type='C' and nodeis_active = true order by node_host;"
	}
	rows, err := db.Query(queryStr)
	if err != nil {
		return err
	}

	d.Log(ctx, LogLevelDebug, "Query CN list successfully!", map[string]interface{}{})
	curCNs := make([]coordinateNode, 0, 10)
	rowCount := 1
	for rows.Next() {
		var (
			host string
			port uint16
		)
		if err = rows.Scan(&host, &port); err != nil {
			d.Log(ctx, LogLevelError,
				fmt.Sprintf("Failed to parse the query result, error info: %v", err.Error()), map[string]interface{}{})
			continue
		}
		curCNs = append(curCNs, coordinateNode{
			ip:   host,
			port: port,
		})
		d.Log(ctx, LogLevelDebug,
			fmt.Sprintf("[Received CN] %v - %v:%v", rowCount, host, port),
			map[string]interface{}{})
		rowCount++
	}
	if len(curCNs) > 0 {
		if err = d.setCNs(curCNs); err != nil {
			return fmt.Errorf("cannot set cn list: %v", err)
		}
	} else {
		d.Log(ctx, LogLevelWarn, "No result is returned for querying the CN list.", map[string]interface{}{})
	}
	return nil
}

func connectCNodeConfig(ctx context.Context, cfg *Config, cNode coordinateNode) (*conn, error) {
	bckCfg := &FallbackConfig{
		Host:      cNode.ip,
		Port:      cNode.port,
		TLSConfig: cNode.TLSConfig,
	}
	cn := &conn{
		config:         cfg,
		logLevel:       cfg.LogLevel,
		logger:         cfg.Logger,
		fallbackConfig: bckCfg,
	}
	cn.log(ctx, LogLevelInfo,
		fmt.Sprintf("Dialing server: (%v:%v)", bckCfg.Host, bckCfg.Port),
		map[string]interface{}{})

	network, address := NetworkAddress(bckCfg.Host, bckCfg.Port)
	var err error
	cn.c, err = cfg.DialFunc(ctx, network, address) // exactly establish connection
	if err != nil {
		return nil, &connectError{config: cfg, msg: fmt.Sprintf("dial error: %v", err), err: driver.ErrBadConn}
	}
	if bckCfg.TLSConfig != nil {
		if err = cn.startTLS(bckCfg.TLSConfig); err != nil {
			if err = cn.c.Close(); err != nil {
				return nil, fmt.Errorf("cannot close connect: %w", err)
			}
			return nil, &connectError{config: cfg, msg: "tls error", err: err}
		}
	}

	cn.buf = bufio.NewReader(cn.c)
	if err = cn.startup(); err != nil {
		_ = cn.Close()
		return nil, fmt.Errorf("fail to startup: %w", err)
	}

	// reset the deadline, in case one was set (see dial)
	if cfg.ConnectTimeout.Seconds() > 0 {
		if err = cn.c.SetDeadline(time.Time{}); err != nil {
			_ = cn.Close()
			return nil, fmt.Errorf("cannot set deadline: %w", err)
		}
	}
	return cn, nil
}

func (d *distributeDialer) setCNs(cns []coordinateNode) error { // TODO:
	if len(cns) == 0 {
		return errors.New("no CN input")
	}
	d.cnsLock.Lock()
	if len(d.coordinateNodes) == 0 {
		d.coordinateNodes = make([]coordinateNode, 0, len(cns))
	} else {
		d.coordinateNodes = d.coordinateNodes[0:0]
	}
	for _, cn := range cns {
		d.coordinateNodes = append(d.coordinateNodes, coordinateNode{
			ip:   cn.ip,
			port: cn.port,
		})
	}
	d.cnsLock.Unlock()
	return nil
}

func (d *distributeDialer) deepCopyCns() []coordinateNode {
	cns := make([]coordinateNode, len(d.coordinateNodes))
	d.cnsLock.RLock()
	copy(cns, d.coordinateNodes)
	defer d.cnsLock.RUnlock()

	return cns
}

func expandWithIPsDist(ctx context.Context, lookupFn LookupFunc, cNodes []coordinateNode) ([]coordinateNode, error) {
	var configs []coordinateNode

	for _, cn := range cNodes {
		// skip resolve for unix sockets
		if strings.HasPrefix(cn.ip, "/") {
			configs = append(configs, coordinateNode{
				ip:        cn.ip,
				port:      cn.port,
				TLSConfig: cn.TLSConfig,
			})
			continue
		}
		ips, err := lookupFn(ctx, cn.ip)
		if err != nil {
			return nil, &connectError{
				config: nil,
				msg:    fmt.Sprintf("fail to look up: %v", err),
				err:    driver.ErrBadConn,
			}
		}
		for _, ip := range ips {
			configs = append(configs, coordinateNode{
				ip:        ip,
				port:      cn.port,
				TLSConfig: cn.TLSConfig,
			})
		}
	}
	return configs, nil
}

type coordinateNode struct {
	ip        string
	port      uint16
	TLSConfig *tls.Config // TODO: separate TLSConfig to dialNode
}

type cnsBalancer interface {
	balance(cns []coordinateNode) int
}

// roundRobbinBalancer
type roundRobbinBalancer struct {
	startIdx int
	idxLock  *sync.Mutex
	*shuffleBalancer
}

func (r *roundRobbinBalancer) balance(cns []coordinateNode) int {
	if len(cns) < 2 {
		return 0
	}
	r.idxLock.Lock()
	roundIdx := r.startIdx
	idx := roundIdx % len(cns)
	r.startIdx++
	r.idxLock.Unlock()
	sufCns := cns[:idx]
	preCns := cns[idx:]
	preCns = append(preCns, sufCns...)
	r.shuffleBalancer.balance(preCns[1:])
	copy(cns, preCns)
	return roundIdx
}

type shuffleBalancer struct{}

func (s *shuffleBalancer) balance(cns []coordinateNode) int {
	rand.Shuffle(len(cns), func(i, j int) {
		cns[i], cns[j] = cns[j], cns[i]
	})
	return 0
}

type priorityBalancer struct {
	num    int
	urlCNs []coordinateNode
	*roundRobbinBalancer
}

func (p *priorityBalancer) balance(cns []coordinateNode) int {
	// Filter out failed cn.
	cnsMap := make(map[string]struct{})
	for i, _ := range cns {
		cnsMap[fmt.Sprintf("%s:%d", cns[i].ip, cns[i].port)] = struct{}{}
	}
	priCNs := make([]coordinateNode, 0, len(cns))
	for i, _ := range p.urlCNs {
		if i >= p.num {
			break
		}
		if _, ok := cnsMap[fmt.Sprintf("%s:%d", p.urlCNs[i].ip, p.urlCNs[i].port)]; ok {
			priCNs = append(priCNs, p.urlCNs[i])
		}
	}
	activeUrlCNNum := len(priCNs)
	idx := p.roundRobbinBalancer.balance(priCNs)

	// add non-priority CN from cns
	m := make(map[string]struct{})
	for _, cn := range priCNs {
		m[fmt.Sprintf("%s:%d", cn.ip, cn.port)] = struct{}{}
	}
	for i := 0; i < len(cns); i++ {
		if _, ok := m[fmt.Sprintf("%s:%d", cns[i].ip, cns[i].port)]; !ok {
			priCNs = append(priCNs, cns[i])
		}
	}
	p.shuffleBalancer.balance(priCNs[activeUrlCNNum:])
	copy(cns, priCNs)
	return idx
}
