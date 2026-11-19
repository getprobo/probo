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

const taskListFilterParams = ["q", "status", "assignee", "due", "internalControl"] as const;

export function taskListHasFilters(params: URLSearchParams): boolean {
  return taskListFilterParams.some(key => params.has(key));
}

// Returns null when the URL already has filters, or when there is no profile
// to assign. Otherwise adds the current user and keeps every other param.
export function myTasksSearchParams(
  params: URLSearchParams,
  profileId: string,
): URLSearchParams | null {
  if (profileId === "" || taskListHasFilters(params)) {
    return null;
  }

  const next = new URLSearchParams(params);
  next.set("assignee", profileId);
  return next;
}
