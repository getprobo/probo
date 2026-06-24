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

const sizes = {
  default: { track: { width: 44, height: 24, padding: 2 }, thumb: 20 },
  sm: { track: { width: 32, height: 18, padding: 2 }, thumb: 14 },
} as const;

type Props = {
  checked: boolean;
  onChange: (checked: boolean) => void;
  disabled?: boolean;
  size?: keyof typeof sizes;
  title?: string;
};

export function Toggle({ checked, onChange, disabled = false, size = "default", title }: Props) {
  const { track, thumb } = sizes[size];
  const travel = track.width - thumb - track.padding * 2;

  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      disabled={disabled}
      title={title}
      onClick={() => !disabled && onChange(!checked)}
      style={{
        position: "relative",
        display: "inline-flex",
        alignItems: "center",
        flexShrink: 0,
        width: track.width,
        height: track.height,
        padding: track.padding,
        borderRadius: 9999,
        border: "none",
        cursor: disabled ? "not-allowed" : "pointer",
        opacity: disabled ? 0.5 : 1,
        backgroundColor: checked
          ? "var(--color-accent)"
          : "var(--color-border-mid)",
        transition: "background-color 200ms ease-in-out",
      }}
    >
      <span
        style={{
          display: "block",
          width: thumb,
          height: thumb,
          borderRadius: 9999,
          backgroundColor: "white",
          boxShadow: "0 1px 2px rgba(0,0,0,0.1)",
          transition: "transform 200ms ease-in-out",
          transform: checked ? `translateX(${travel}px)` : "translateX(0)",
        }}
      />
    </button>
  );
}
