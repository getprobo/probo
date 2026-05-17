import type { ComponentProps } from "react";

export function Monday(props: ComponentProps<"svg">) {
  return (
    <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" {...props}>
      <path
        d="M2.25 16.5L6.75 6a2.25 2.25 0 014.13 1.78L6.38 18.28A2.25 2.25 0 012.25 16.5zm7.5 0L14.25 6a2.25 2.25 0 014.13 1.78L13.88 18.28A2.25 2.25 0 019.75 16.5zM18 18.75a2.25 2.25 0 110-4.5 2.25 2.25 0 010 4.5z"
        fill="#FF3D57"
      />
    </svg>
  );
}
