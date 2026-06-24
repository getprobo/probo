import type { ComponentProps } from "react";

export function IncidentIO(props: ComponentProps<"svg">) {
  return (
    <svg viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg" {...props}>
      <rect
        x="42"
        y="19.5"
        width="9"
        height="36"
        transform="rotate(90 42 19.5)"
        fill="#F25533"
      />
      <rect x="19.5" y="6" width="9" height="36" fill="#F25533" />
    </svg>
  );
}
