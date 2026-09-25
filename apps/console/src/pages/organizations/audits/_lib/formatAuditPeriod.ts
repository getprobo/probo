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

interface FormatAuditPeriodOptions {
  language: string;
  start?: string | null;
  end?: string | null;
  from: (date: string) => string;
  until: (date: string) => string;
  notSet: string;
}

const compactDateOptions: Intl.DateTimeFormatOptions = {
  year: "numeric",
  month: "short",
  day: "numeric",
};

const datetimeSuffixPattern
  = /^T(?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d(?:\.\d+)?(?:Z|[+-](?:[01]\d|2[0-3]):[0-5]\d)$/;

function parsePeriodDate(value?: string | null): Date | null {
  const match = value?.match(/^(\d{4})-(\d{2})-(\d{2})(.*)$/);
  if (!match || (match[4] && !datetimeSuffixPattern.test(match[4]))) {
    return null;
  }

  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3]);
  const date = new Date(year, month - 1, day);

  if (
    date.getFullYear() !== year
    || date.getMonth() !== month - 1
    || date.getDate() !== day
  ) {
    return null;
  }

  return date;
}

export function formatAuditPeriod({
  language,
  start,
  end,
  from,
  until,
  notSet,
}: FormatAuditPeriodOptions): string {
  const startDate = parsePeriodDate(start);
  const endDate = parsePeriodDate(end);
  const formatter = new Intl.DateTimeFormat(language, compactDateOptions);

  if (startDate && endDate) {
    if (startDate < endDate) {
      return formatter.formatRange(startDate, endDate);
    }

    if (startDate.getTime() === endDate.getTime()) {
      return formatter.format(startDate);
    }

    return `${formatter.format(startDate)} – ${formatter.format(endDate)}`;
  }
  if (startDate) {
    return from(formatter.format(startDate));
  }
  if (endDate) {
    return until(formatter.format(endDate));
  }
  return notSet;
}
