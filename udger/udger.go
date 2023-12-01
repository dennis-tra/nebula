package udger

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"math/big"
	"net"

	"github.com/friendsofgo/errors"
	_ "github.com/mattn/go-sqlite3"
)

type Client struct {
	db *sql.DB
}

// NewClient initializes a new maxmind database client from the embedded database
func NewClient(dbpath string) (*Client, error) {
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		return nil, errors.Wrap(err, "open udger db")
	}

	return &Client{db: db}, nil
}

func (c *Client) Datacenter(addr string) (int, error) {
	ip := net.ParseIP(addr)
	if ip.To4() != nil {
		return c.handleIPv4(ip)
	} else {
		return c.handleIPv6(ip)
	}
}

func (c *Client) handleIPv4(ip net.IP) (int, error) {
	ipint := big.NewInt(0)
	ipint.SetBytes(ip.To4())
	packedIp, err := packIpInt(ipint.Int64())
	if err != nil {
		return 0, errors.Wrap(err, "pack ip")
	}

	rows, err := c.db.Query("SELECT datacenter_id FROM udger_datacenter_range WHERE iplong_from <= ? AND ? <= iplong_to", packedIp, packedIp)
	if err != nil {
		return 0, errors.Wrap(err, "query ip4 datacenter")
	} else if rows.Err() != nil {
		return 0, errors.Wrap(err, "query ip4 datacenter")
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, sql.ErrNoRows
	}

	var datacenter_id int
	if err = rows.Scan(&datacenter_id); err != nil {
		return 0, errors.Wrap(err, "scan ip4 datacenter query")
	}

	return datacenter_id, nil
}

func (c *Client) handleIPv6(ip net.IP) (int, error) {
	comps := make([]int64, 8)
	for i := 0; i < 8; i += 2 {
		ipint := big.NewInt(0)
		ipint.SetBytes(ip.To16()[i : i+2])
		packedIp, err := packIpInt(ipint.Int64())
		if err != nil {
			return 0, errors.Wrap(err, "pack ip")
		}
		comps[i/2] = packedIp
	}

	rows, err := c.db.Query(`
	SELECT datacenter_id FROM udger_datacenter_range6
	WHERE
		iplong_from0 <= ? AND iplong_from1 <= ? AND iplong_from2 <= ? AND iplong_from3 <= ? AND
		iplong_from4 <= ? AND iplong_from5 <= ? AND iplong_from6 <= ? AND iplong_from7 <= ? AND
		iplong_to0 >= ? AND iplong_to1 >= ? AND iplong_to2 >= ? AND iplong_to3 >= ? AND
		iplong_to4 >= ? AND iplong_to5 >= ? AND iplong_to6 >= ? AND iplong_to7 >= ?
	`, comps[0], comps[1], comps[2], comps[3], comps[4], comps[5], comps[6], comps[7],
		comps[0], comps[1], comps[2], comps[3], comps[4], comps[5], comps[6], comps[7])
	if err != nil {
		return 0, errors.Wrap(err, "query ip6 datacenter")
	} else if rows.Err() != nil {
		return 0, errors.Wrap(err, "query ip6 datacenter")
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, sql.ErrNoRows
	}

	var datacenter_id int
	if err = rows.Scan(&datacenter_id); err != nil {
		return 0, errors.Wrap(err, "scan ip6 datacenter query")
	}

	return datacenter_id, nil
}

func packIpInt(ipint int64) (int64, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, uint32(ipint))
	if err != nil {
		return 0, errors.Wrap(err, "write ip int to buffer")
	}
	packedInt := big.NewInt(0)
	packedInt.SetBytes(buf.Bytes())
	return packedInt.Int64(), nil
}
