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

import type { SubdivisionPath } from "./worldMap";
import { WORLD_VIEW_BOX, worldCountryPaths } from "./worldMap";

export interface SubdivisionMap {
  paths: SubdivisionPath[];
  activeId: string;
  viewBox: string;
}

export function wantsSubdivisionMap(
  countryCode: string | null | undefined,
  subdivisionCode: string | null | undefined,
): boolean {
  return (countryCode === "US" || countryCode === "CA")
    && subdivisionCode != null
    && subdivisionCode !== "";
}

export function locationMapSvg(
  subdivision: SubdivisionMap | null,
  loading: boolean,
  countryId: string | null,
  mapSvg: string,
  country: string,
  countryActive: string,
) {
  if (loading) {
    return <svg className={mapSvg} viewBox={WORLD_VIEW_BOX} aria-hidden />;
  }
  if (subdivision != null) {
    return (
      <svg className={mapSvg} viewBox={subdivision.viewBox} aria-hidden>
        {subdivision.paths.map(path => (
          <path
            key={path.id}
            d={path.d}
            className={path.id === subdivision.activeId ? countryActive : country}
            fillRule="evenodd"
            strokeWidth={0.5}
          />
        ))}
      </svg>
    );
  }
  return (
    <svg className={mapSvg} viewBox={WORLD_VIEW_BOX} aria-hidden>
      {worldCountryPaths().map(path => (
        <path
          key={path.id}
          d={path.d}
          className={path.id === countryId ? countryActive : country}
          fillRule="evenodd"
          strokeWidth={0.5}
        />
      ))}
    </svg>
  );
}
