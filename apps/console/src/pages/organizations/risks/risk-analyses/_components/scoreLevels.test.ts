import { describe, expect, it } from "vitest";

import { scoreLevelRanges } from "./scoreLevels";

describe("scoreLevelRanges", () => {
  it("ends each range before the next level and the last range at the matrix maximum", () => {
    expect(scoreLevelRanges(
      [
        { level: "LOW", minScore: 1 },
        { level: "HIGH", minScore: 13 },
      ],
      25,
    )).toEqual([
      { level: "LOW", range: "1-12" },
      { level: "HIGH", range: "13-25" },
    ]);
  });
});
