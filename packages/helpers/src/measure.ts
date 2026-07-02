// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

type Translator = (s: string) => string;

export const measureStates = [
    "IMPLEMENTED",
    "IN_PROGRESS",
    "NOT_APPLICABLE",
    "NOT_STARTED",
    "UNKNOWN",
    "NOT_IMPLEMENTED",
] as const;

export function getMeasureStateLabel(__: Translator, state: string) {
    switch (state) {
        case "IMPLEMENTED":
            return __("Implemented");
        case "IN_PROGRESS":
            return __("In Progress");
        case "NOT_APPLICABLE":
            return __("Not Applicable");
        case "NOT_STARTED":
            return __("Not Started");
        case "UNKNOWN":
            return __("Unknown");
        case "NOT_IMPLEMENTED":
            return __("Not Implemented");
        default:
            return __("Unknown");
    }
}
