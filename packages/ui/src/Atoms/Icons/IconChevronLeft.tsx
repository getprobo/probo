import type { IconProps } from "./type";

export function IconChevronLeft({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M15.8485 3.35147C16.3172 3.8201 16.3172 4.5799 15.8485 5.04853L8.89706 12L15.8485 18.9515C16.3172 19.4201 16.3172 20.1799 15.8485 20.6485C15.3799 21.1172 14.6201 21.1172 14.1515 20.6485L6.35147 12.8485C5.88284 12.3799 5.88284 11.6201 6.35147 11.1515L14.1515 3.35147C14.6201 2.88284 15.3799 2.88284 15.8485 3.35147Z" fill="currentColor" />
    </svg>
  );
}
