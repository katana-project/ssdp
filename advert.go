// Copyright 2014 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssdp

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"net/url"
)

func newAdvert(method, host string, hdr http.Header) *http.Request {
	return &http.Request{
		Method:     method,
		URL:        &url.URL{Path: "*"},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     hdr,
		Host:       host,
	}
}

func parseAdvert(b []byte) (*http.Request, error) {
	return http.ReadRequest(bufio.NewReaderSize(bytes.NewReader(b), len(b)))
}

// An AdvertRedirector represents a SSDP advertisement message
// redirector.
type AdvertRedirector struct {
	conn // network connection endpoint
	mifs []net.Interface
	path *path // reverse path
	req  *http.Request
}

// Header returns the HTTP header map that will be sent by WriteTo
// method.
func (rdr *AdvertRedirector) Header() http.Header {
	return rdr.req.Header
}

// WriteTo writes the SSDP advertisement message. The outbound network
// interface ifi is used for sending multicast message. It uses the
// system assigned multicast network interface when ifi is nil.
func (rdr *AdvertRedirector) WriteTo(dst *net.UDPAddr, ifi *net.Interface) (int, error) {
	if ifi != nil {
		rdr.SetMulticastInterface(ifi)
	}
	var buf bytes.Buffer
	if err := rdr.req.Write(&buf); err != nil {
		return 0, err
	}
	return rdr.writeTo(buf.Bytes(), dst)
}

// ForwardPath returns the destination address of the SSDP
// advertisement message.
func (rdr *AdvertRedirector) ForwardPath() *net.UDPAddr {
	return rdr.path.dst
}

// ReversePath returns the source address and inbound interface of the
// SSDP advertisement message.
func (rdr *AdvertRedirector) ReversePath() (*net.UDPAddr, *net.Interface) {
	return rdr.path.src, interfaceByIndex(rdr.mifs, rdr.path.ifIndex)
}

func newAdvertRedirector(conn conn, mifs []net.Interface, grp *net.UDPAddr, path *path, req *http.Request) *AdvertRedirector {
	rdr := &AdvertRedirector{
		conn: conn,
		mifs: mifs,
		path: path,
		req:  req,
	}
	path.dst.Port = grp.Port
	if ipv6LinkLocal(path.src.IP) {
		path.src.Zone = interfaceByIndex(mifs, path.ifIndex).Name
	}
	return rdr
}
