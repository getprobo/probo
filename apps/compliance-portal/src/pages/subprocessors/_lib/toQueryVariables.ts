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

import type { SubprocessorsPageQuery$variables } from "../__generated__/SubprocessorsPageQuery.graphql";

interface FilterValues {
  query: string;
  category: string;
  country: string;
}

// Maps the string-based URL filter state to the typed GraphQL query variables,
// treating empty strings as "no filter". The category/country strings originate
// from the enum-constrained Select options, so the cast is safe.
export function toQueryVariables(filters: FilterValues): SubprocessorsPageQuery$variables {
  return {
    query: filters.query || null,
    category: (filters.category || null) as SubprocessorsPageQuery$variables["category"],
    country: (filters.country || null) as SubprocessorsPageQuery$variables["country"],
  };
}
