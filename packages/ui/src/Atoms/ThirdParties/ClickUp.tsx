import type { ComponentProps } from "react";

export function ClickUp(props: ComponentProps<"svg">) {
  return (
    <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" {...props}>
      <path
        d="m2 18.439 3.69-2.828c1.961 2.56 4.044 3.739 6.363 3.739 2.307 0 4.33-1.166 6.203-3.708L22 18.405C19.298 22.067 15.941 24 12.053 24 8.178 24 4.788 22.08 2 18.439Z"
        fill="#7B68EE"
      />
      <path
        d="m12.04 7.476-6.566 5.66-3.035-3.519L12.054 1.36l9.543 8.265-3.05 3.508z"
        fill="#7B68EE"
      />
    </svg>
  );
}
