// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package mcp_v1

import (
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.probo.inc/probo/pkg/server/api/clientip"
)

type signerMetadata struct {
	IPAddr string
	UA     string
}

func signerMetadataFromToolRequest(req *mcp.CallToolRequest) signerMetadata {
	if req == nil || req.Extra == nil || req.Extra.Header == nil {
		return signerMetadata{}
	}

	httpReq := &http.Request{Header: req.Extra.Header}

	return signerMetadata{
		IPAddr: clientip.Extract(httpReq),
		UA:     req.Extra.Header.Get("User-Agent"),
	}
}
