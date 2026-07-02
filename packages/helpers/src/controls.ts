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

type Translator = (s: string) => string;

export const controlMaturityLevels = [
    "NONE",
    "INITIAL",
    "MANAGED",
    "DEFINED",
    "QUANTITATIVELY_MANAGED",
    "OPTIMIZING",
] as const;

export type ControlMaturityLevel = (typeof controlMaturityLevels)[number];

export function getControlMaturityLevelLabel(__: Translator, level: string) {
    switch (level) {
        case "NONE":
            return __("0 - None");
        case "INITIAL":
            return __("1 - Initial");
        case "MANAGED":
            return __("2 - Managed");
        case "DEFINED":
            return __("3 - Defined");
        case "QUANTITATIVELY_MANAGED":
            return __("4 - Quantitatively Managed");
        case "OPTIMIZING":
            return __("5 - Optimizing");
    }
}
