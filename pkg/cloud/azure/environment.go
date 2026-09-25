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

package azure

import (
	"encoding"
	"fmt"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
)

// Environment is the Azure cloud a session dials. DoD shares an authority
// with Government and differs only on the Graph host.
type Environment string

const (
	EnvironmentPublic        Environment = "AZURE_PUBLIC"
	EnvironmentGovernment    Environment = "AZURE_GOVERNMENT"
	EnvironmentGovernmentDoD Environment = "AZURE_GOVERNMENT_DOD"
	EnvironmentChina         Environment = "AZURE_CHINA"

	graphDefaultPath = ".default"
)

type endpoints struct {
	configuration cloud.Configuration
	graphBaseURL  string
	graphScope    string
	armScope      string
}

var (
	_ fmt.Stringer             = Environment("")
	_ encoding.TextMarshaler   = Environment("")
	_ encoding.TextUnmarshaler = (*Environment)(nil)

	environmentTable = map[Environment]endpoints{
		EnvironmentPublic: {
			configuration: cloud.AzurePublic,
			graphBaseURL:  "https://graph.microsoft.com",
			graphScope:    mustScope("https://graph.microsoft.com"),
			armScope:      mustScope(cloud.AzurePublic.Services[cloud.ResourceManager].Audience),
		},
		EnvironmentGovernment: {
			configuration: cloud.AzureGovernment,
			graphBaseURL:  "https://graph.microsoft.us",
			graphScope:    mustScope("https://graph.microsoft.us"),
			armScope:      mustScope(cloud.AzureGovernment.Services[cloud.ResourceManager].Audience),
		},
		EnvironmentGovernmentDoD: {
			configuration: cloud.AzureGovernment,
			graphBaseURL:  "https://dod-graph.microsoft.us",
			graphScope:    mustScope("https://dod-graph.microsoft.us"),
			armScope:      mustScope(cloud.AzureGovernment.Services[cloud.ResourceManager].Audience),
		},
		EnvironmentChina: {
			configuration: cloud.AzureChina,
			graphBaseURL:  "https://microsoftgraph.chinacloudapi.cn",
			graphScope:    mustScope("https://microsoftgraph.chinacloudapi.cn"),
			armScope:      mustScope(cloud.AzureChina.Services[cloud.ResourceManager].Audience),
		},
	}
)

func Environments() []Environment {
	return []Environment{
		EnvironmentPublic,
		EnvironmentGovernment,
		EnvironmentGovernmentDoD,
		EnvironmentChina,
	}
}

func (v Environment) IsValid() bool {
	switch v {
	case
		EnvironmentPublic,
		EnvironmentGovernment,
		EnvironmentGovernmentDoD,
		EnvironmentChina:
		return true
	}

	return false
}

func (v Environment) String() string {
	return string(v)
}

func (v Environment) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *Environment) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*v = EnvironmentPublic
		return nil
	}

	val := Environment(text)
	if !val.IsValid() {
		return errInvalidEnvironment
	}

	*v = val

	return nil
}

func (v Environment) endpoints() (endpoints, error) {
	if v == "" {
		v = EnvironmentPublic
	}

	ep, ok := environmentTable[v]
	if !ok {
		return endpoints{}, errInvalidEnvironment
	}

	return ep, nil
}

func mustScope(base string) string {
	scope, err := url.JoinPath(base, graphDefaultPath)
	if err != nil {
		panic(err)
	}

	return scope
}
