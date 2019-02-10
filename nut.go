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

	fmt.Println("readall")
	data, err := c.listVarsResponse()

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

	lines := strings.Split(text, "\n")
	for _, s := range lines {
		fmt.Println(s)
		bits := strings.SplitN(s, ": ", 2)
		if len(bits) != 2 {
			continue
		}
		k := bits[0]
		v := bits[1]

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

		collected += s
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", nil
}
