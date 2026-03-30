// --license-header-ignore
// Microsoft logo — trademark of Microsoft Corporation
import type { ComponentProps } from "react";

export function Microsoft(props: ComponentProps<"svg">) {
  return (
    <svg
      width="800px"
      height="800px"
      viewBox="0 0 256 256"
      xmlns="http://www.w3.org/2000/svg"
      preserveAspectRatio="xMidYMid"
      {...props}
    >
      <rect x="0" y="0" width="121" height="121" fill="#F25022" />
      <rect x="135" y="0" width="121" height="121" fill="#7FBA00" />
      <rect x="0" y="135" width="121" height="121" fill="#00A4EF" />
      <rect x="135" y="135" width="121" height="121" fill="#FFB900" />
    </svg>
  );
}
