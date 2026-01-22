import type { IconProps } from "./type";

export function IconChevronUp({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M11.1515 6.35147C11.6201 5.88284 12.3799 5.88284 12.8485 6.35147L20.6485 14.1515C21.1172 14.6201 21.1172 15.3799 20.6485 15.8485C20.1799 16.3172 19.4201 16.3172 18.9515 15.8485L12 8.89706L5.04853 15.8485C4.5799 16.3172 3.8201 16.3172 3.35147 15.8485C2.88284 15.3799 2.88284 14.6201 3.35147 14.1515L11.1515 6.35147Z" fill="currentColor" />
    </svg>
  );
}
