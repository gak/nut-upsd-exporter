package nut

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type UPSClient struct {
	HostIP string
	Name   string

	conn net.Conn
}

type Results map[string]string

func (c *UPSClient) All() (Results, error) {
	fmt.Println("connecting")
	conn, err := c.Connect()
	if err != nil {
		return nil, err
	}

	c.conn = conn

	fmt.Println("listvars")
	err = c.listVars()
	if err != nil {
		return nil, err
	}

	fmt.Println("list vars response")
	data, err := c.listVarsResponse()

	fmt.Println("parse")
	return c.parse(data)
}

func (c *UPSClient) Connect() (net.Conn, error) {
	return net.Dial("tcp", c.HostIP)
}

func (c *UPSClient) listVars() error {
	_, err := fmt.Fprintf(c.conn, "LIST VAR %s\n", c.Name)
	return err
}

func (c *UPSClient) parse(text string) (Results, error) {
	r := Results{}

	// VAR eaton key "value"

	lines := strings.Split(text, "\n")
	for _, s := range lines {
		bits := strings.SplitN(s, " ", 4)
		if len(bits) != 4 {
			continue
		}

		k := bits[2]
		v := bits[3]
		// Strip off the start and end quotes
		v = v[1 : len(v)-1]

		r[k] = v
	}

	return r, nil
}

func (c *UPSClient) listVarsResponse() (string, error) {
	scanner := bufio.NewScanner(c.conn)
	collected := ""
	for scanner.Scan() {
		s := scanner.Text()
		if strings.HasPrefix(s, "END LIST VAR") {
			return collected, nil
		}

		collected += s + "\n"
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", nil
}
