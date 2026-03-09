const SOCIAL_PATTERNS: Array<[RegExp, string]> = [
  [/linkedin\.com/i, "LinkedIn"],
  [/(?:twitter|x)\.com/i, "X"],
  [/facebook\.com/i, "Facebook"],
];

export function detectSocialName(url: string): string | null {
  for (const [pattern, name] of SOCIAL_PATTERNS) {
    if (pattern.test(url)) return name;
  }
  return null;
}
