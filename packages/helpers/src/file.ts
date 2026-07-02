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

/**
 * Return the file size in a human readable format
 */
export function fileSize(__: (s: string) => string, size: number): string {
    if (size < 0) return "";
    if (size === 0) return `0 ${__("B")}`;

    const units = [__("B"), __("KB"), __("MB"), __("GB"), __("TB")];
    const i = Math.floor(Math.log(size) / Math.log(1024));

    // Don't go beyond available units
    const unitIndex = Math.min(i, units.length - 1);

    // Convert to the appropriate unit with 2 decimal places
    const convertedSize = size / Math.pow(1024, unitIndex);
    const formattedSize = Math.round(convertedSize * 100) / 100;

    return `${formattedSize} ${units[unitIndex]}`;
}

type FileInfo = {
    type: string;
    mimeType: string;
};

export function fileType(__: (s: string) => string, info: FileInfo): string {
    if (
        info.type !== "FILE" ||
        (info.mimeType !== "text/uri-list" && info.mimeType !== "text/uri")
    ) {
        return __("Document");
    }
    return __("Link");
}
