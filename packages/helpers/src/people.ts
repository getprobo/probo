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

export const peopleRoles = [
    "EMPLOYEE",
    "CONTRACTOR",
    "SERVICE_ACCOUNT",
] as const;

export function getRoles(__: Translator) {
    return [
        {
            value: "EMPLOYEE",
            label: __("Employee"),
        },
        {
            value: "CONTRACTOR",
            label: __("Contractor"),
        },
        {
            value: "SERVICE_ACCOUNT",
            label: __("Service account"),
        },
    ];
}

export function getRole(__: Translator, role?: string): string {
    switch (role) {
        case "EMPLOYEE":
            return __("Employee");
        case "CONTRACTOR":
            return __("Contractor");
        case "SERVICE_ACCOUNT":
            return __("Service account");
        default:
            return __("Unknown");
    }
}
