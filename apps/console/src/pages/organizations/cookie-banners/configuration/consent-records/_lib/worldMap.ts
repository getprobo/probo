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

import type { Feature, MultiPolygon, Polygon, Position } from "geojson";
import { feature } from "topojson-client";
import type { GeometryCollection, Topology } from "topojson-specification";
import countriesTopology from "world-atlas/countries-110m.json";

export const WORLD_MAP_WIDTH = 960;
export const WORLD_MAP_HEIGHT = 480;

export const WORLD_VIEW_BOX = `0 0 ${WORLD_MAP_WIDTH} ${WORLD_MAP_HEIGHT}`;

export interface CountryPath {
  id: string;
  d: string;
  viewBox: string | null;
}

export interface SubdivisionPath {
  id: string;
  d: string;
  viewBox: string;
}

export type SubdivisionTopology = Topology<{
  subdivisions: GeometryCollection;
}>;

interface FocusBox {
  minX: number;
  minY: number;
  maxX: number;
  maxY: number;
}

// Extra space around the highlighted region, and a floor so a small
// state still shows its neighbors instead of a cropped blob.
const FOCUS_PAD_RATIO = 0.25;
const FOCUS_MIN_SPAN = 56;

// ISO 3166-1 alpha-2 + 3-digit numeric, concatenated in 5-character groups.
const ISO_3166_1_NUMERIC_PACKED
  = "AD020AE784AF004AG028AI660AL008AM051AO024AQ010AR032AS016AT040"
    + "AU036AW533AX248AZ031BA070BB052BD050BE056BF854BG100BH048BI108"
    + "BJ204BL652BM060BN096BO068BQ535BR076BS044BT064BV074BW072BY112"
    + "BZ084CA124CC166CD180CF140CG178CH756CI384CK184CL152CM120CN156"
    + "CO170CR188CU192CV132CW531CX162CY196CZ203DE276DJ262DK208DM212"
    + "DO214DZ012EC218EE233EG818EH732ER232ES724ET231FI246FJ242FK238"
    + "FM583FO234FR250GA266GB826GD308GE268GF254GG831GH288GI292GL304"
    + "GM270GN324GP312GQ226GR300GS239GT320GU316GW624GY328HK344HM334"
    + "HN340HR191HT332HU348ID360IE372IL376IM833IN356IO086IQ368IR364"
    + "IS352IT380JE832JM388JO400JP392KE404KG417KH116KI296KM174KN659"
    + "KP408KR410KW414KY136KZ398LA418LB422LC662LI438LK144LR430LS426"
    + "LT440LU442LV428LY434MA504MC492MD498ME499MF663MG450MH584MK807"
    + "ML466MM104MN496MO446MP580MQ474MR478MS500MT470MU480MV462MW454"
    + "MX484MY458MZ508NA516NC540NE562NF574NG566NI558NL528NO578NP524"
    + "NR520NU570NZ554OM512PA591PE604PF258PG598PH608PK586PL616PM666"
    + "PN612PR630PS275PT620PW585PY600QA634RE638RO642RS688RU643RW646"
    + "SA682SB090SC690SD729SE752SG702SH654SI705SJ744SK703SL694SM674"
    + "SN686SO706SR740SS728ST678SV222SX534SY760SZ748TC796TD148TF260"
    + "TG768TH764TJ762TK772TL626TM795TN788TO776TR792TT780TV798TW158"
    + "TZ834UA804UG800UM581US840UY858UZ860VA336VC670VE862VG092VI850"
    + "VN704VU548WF876WS882YE887YT175ZA710ZM894ZW716";

const isoNumericByAlpha2 = new Map<string, string>();

for (let i = 0; i < ISO_3166_1_NUMERIC_PACKED.length; i += 5) {
  isoNumericByAlpha2.set(
    ISO_3166_1_NUMERIC_PACKED.slice(i, i + 2),
    ISO_3166_1_NUMERIC_PACKED.slice(i + 2, i + 5),
  );
}

export function isoNumericId(alpha2: string): string | null {
  return isoNumericByAlpha2.get(alpha2.toUpperCase()) ?? null;
}

function project(lon: number, lat: number): [number, number] {
  return [
    ((lon + 180) / 360) * WORLD_MAP_WIDTH,
    ((90 - lat) / 180) * WORLD_MAP_HEIGHT,
  ];
}

function ringPath(ring: Position[]): string {
  return `${ring.map((position, index) => {
    const [x, y] = project(position[0] ?? 0, position[1] ?? 0);
    return `${index === 0 ? "M" : "L"}${x.toFixed(1)} ${y.toFixed(1)}`;
  }).join(" ")}Z`;
}

function geometryPath(geometry: Polygon | MultiPolygon): string {
  if (geometry.type === "Polygon") {
    return geometry.coordinates.map(ringPath).join("");
  }
  return geometry.coordinates.flatMap(polygon => polygon.map(ringPath)).join("");
}

function polygonFocus(rings: Position[][]): FocusBox | null {
  let minLon = Infinity;
  let maxLon = -Infinity;
  let minX = Infinity;
  let minY = Infinity;
  let maxX = -Infinity;
  let maxY = -Infinity;

  for (const ring of rings) {
    for (const position of ring) {
      const lon = position[0] ?? 0;
      const lat = position[1] ?? 0;
      const [x, y] = project(lon, lat);
      minLon = Math.min(minLon, lon);
      maxLon = Math.max(maxLon, lon);
      minX = Math.min(minX, x);
      minY = Math.min(minY, y);
      maxX = Math.max(maxX, x);
      maxY = Math.max(maxY, y);
    }
  }

  if (!Number.isFinite(minX) || maxLon - minLon >= 180) {
    return null;
  }

  return { minX, minY, maxX, maxY };
}

// Overseas pieces sit in the same feature. Framing the largest piece
// keeps Alaska or Hawaii from pulling a California view across the map.
function geometryFocus(geometry: Polygon | MultiPolygon): FocusBox | null {
  const polygons = geometry.type === "Polygon"
    ? [geometry.coordinates]
    : geometry.coordinates;
  let best: FocusBox | null = null;
  let bestArea = -1;

  for (const polygon of polygons) {
    const box = polygonFocus(polygon);
    if (box == null) {
      continue;
    }
    const area = (box.maxX - box.minX) * (box.maxY - box.minY);
    if (area > bestArea) {
      best = box;
      bestArea = area;
    }
  }

  return best;
}

function paddedViewBox(focus: FocusBox): string {
  const spanX = focus.maxX - focus.minX;
  const spanY = focus.maxY - focus.minY;
  const padX = Math.max(spanX * FOCUS_PAD_RATIO, (FOCUS_MIN_SPAN - spanX) / 2);
  const padY = Math.max(spanY * FOCUS_PAD_RATIO, (FOCUS_MIN_SPAN - spanY) / 2);
  const width = Math.min(spanX + padX * 2, WORLD_MAP_WIDTH);
  const height = Math.min(spanY + padY * 2, WORLD_MAP_HEIGHT);
  const x = Math.min(Math.max(0, focus.minX - padX), WORLD_MAP_WIDTH - width);
  const y = Math.min(Math.max(0, focus.minY - padY), WORLD_MAP_HEIGHT - height);

  return `${x.toFixed(1)} ${y.toFixed(1)} ${width.toFixed(1)} ${height.toFixed(1)}`;
}

function countryPath(featureObject: Feature): CountryPath | null {
  if (featureObject.id == null || featureObject.geometry == null) {
    return null;
  }
  const geometry = featureObject.geometry;
  if (geometry.type !== "Polygon" && geometry.type !== "MultiPolygon") {
    return null;
  }
  const d = geometryPath(geometry);
  if (d === "") {
    return null;
  }
  const focus = geometryFocus(geometry);
  return {
    id: String(featureObject.id),
    d,
    viewBox: focus == null ? null : paddedViewBox(focus),
  };
}

function subdivisionPath(featureObject: Feature): SubdivisionPath | null {
  if (featureObject.id == null || featureObject.geometry == null) {
    return null;
  }
  const geometry = featureObject.geometry;
  if (geometry.type !== "Polygon" && geometry.type !== "MultiPolygon") {
    return null;
  }
  const d = geometryPath(geometry);
  const focus = geometryFocus(geometry);
  if (d === "" || focus == null) {
    return null;
  }
  return { id: String(featureObject.id), d, viewBox: paddedViewBox(focus) };
}

const countryPaths: CountryPath[] = feature(
  countriesTopology,
  countriesTopology.objects.countries,
).features.flatMap((featureObject) => {
  const path = countryPath(featureObject);
  return path == null ? [] : [path];
});

export function worldCountryPaths(): CountryPath[] {
  return countryPaths;
}

// Largest-landmass framing crops a real part of these countries
// (date line, archipelagos, or a large sibling like Alaska / Zealand).
const WORLD_FRAME_COUNTRY_IDS = new Set([
  "010", // Antarctica
  "044", // Bahamas
  "090", // Solomon Islands
  "208", // Denmark
  "360", // Indonesia
  "392", // Japan
  "458", // Malaysia
  "548", // Vanuatu
  "554", // New Zealand
  "608", // Philippines
  "643", // Russia
  "840", // United States
]);

export function countryViewBox(id: string | null): string {
  if (id == null || WORLD_FRAME_COUNTRY_IDS.has(id)) {
    return WORLD_VIEW_BOX;
  }
  const country = countryPaths.find(path => path.id === id);
  return country?.viewBox ?? WORLD_VIEW_BOX;
}

export function subdivisionPaths(topology: SubdivisionTopology): SubdivisionPath[] {
  return feature(topology, topology.objects.subdivisions).features.flatMap((featureObject) => {
    const path = subdivisionPath(featureObject);
    return path == null ? [] : [path];
  });
}
