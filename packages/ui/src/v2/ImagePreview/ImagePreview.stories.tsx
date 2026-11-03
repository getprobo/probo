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

import { ImagePreview } from "./ImagePreview";
import { ImagePreviewPopup } from "./ImagePreviewPopup";
import { ImagePreviewSkeleton } from "./ImagePreviewSkeleton";
import { ImagePreviewTrigger } from "./ImagePreviewTrigger";

const src = `data:image/svg+xml,${encodeURIComponent(`
<svg xmlns="http://www.w3.org/2000/svg" width="960" height="540">
  <rect width="960" height="540" fill="#f4f1ea"/>
  <rect x="32" y="32" width="896" height="476" rx="16" fill="#ffffff" stroke="#d8d2c8"/>
  <text x="72" y="120" font-family="sans-serif" font-size="40" fill="#1c1917">Access review evidence</text>
  <text x="72" y="190" font-family="sans-serif" font-size="24" fill="#44403c">Owner: Riley Chen</text>
  <text x="72" y="260" font-family="sans-serif" font-size="22" fill="#1c1917">1. Export the directory snapshot.</text>
  <text x="72" y="310" font-family="sans-serif" font-size="22" fill="#1c1917">2. Compare each account with the approved list.</text>
  <text x="72" y="360" font-family="sans-serif" font-size="22" fill="#1c1917">3. Revoke accounts with no business owner.</text>
</svg>
`)}`;

export default {
  title: "v2/ImagePreview",
  component: ImagePreview,
};

export function Default() {
  return (
    <div className="w-80">
      <ImagePreview>
        <ImagePreviewTrigger aria-label="Enlarge access review evidence">
          <img className="block w-full" src={src} alt="Access review evidence" />
        </ImagePreviewTrigger>
        <ImagePreviewPopup
          src={src}
          alt="Access review evidence"
          label="Enlarge access review evidence"
          closeLabel="Close image"
        />
      </ImagePreview>
    </div>
  );
}

export function Skeleton() {
  return <ImagePreviewSkeleton className="w-80" />;
}
