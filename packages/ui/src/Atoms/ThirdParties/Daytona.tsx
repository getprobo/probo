import type { ComponentProps } from "react";

export function Daytona(props: ComponentProps<"svg">) {
  return (
    <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" {...props}>
      <path
        fill="#F5C400"
        d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm0 3.6A8.4 8.4 0 1 1 3.6 12 8.41 8.41 0 0 1 12 3.6zm-1.8 3.3v10.2h1.95c2.85 0 4.65-1.65 4.65-5.1s-1.8-5.1-4.65-5.1H10.2zm1.8 1.8h.15c1.65 0 2.7.9 2.7 3.3s-1.05 3.3-2.7 3.3H12V8.7z"
      />
    </svg>
  );
}
