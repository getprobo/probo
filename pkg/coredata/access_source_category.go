// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package coredata

import (
	"database/sql/driver"
	"fmt"
)

type AccessSourceCategory string

const (
	AccessSourceCategorySaaS       AccessSourceCategory = "SAAS"
	AccessSourceCategoryCloudInfra AccessSourceCategory = "CLOUD_INFRA"
	AccessSourceCategorySourceCode AccessSourceCategory = "SOURCE_CODE"
	AccessSourceCategoryOther      AccessSourceCategory = "OTHER"
)

func AccessSourceCategories() []AccessSourceCategory {
	return []AccessSourceCategory{
		AccessSourceCategorySaaS,
		AccessSourceCategoryCloudInfra,
		AccessSourceCategorySourceCode,
		AccessSourceCategoryOther,
	}
}

func (c AccessSourceCategory) String() string {
	return string(c)
}

func (c *AccessSourceCategory) Scan(value any) error {
	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("unsupported type for AccessSourceCategory: %T", value)
	}

	switch str {
	case "SAAS":
		*c = AccessSourceCategorySaaS
	case "CLOUD_INFRA":
		*c = AccessSourceCategoryCloudInfra
	case "SOURCE_CODE":
		*c = AccessSourceCategorySourceCode
	case "OTHER":
		*c = AccessSourceCategoryOther
	default:
		return fmt.Errorf("invalid AccessSourceCategory value: %q", str)
	}
	return nil
}

func (c AccessSourceCategory) Value() (driver.Value, error) {
	return c.String(), nil
}
