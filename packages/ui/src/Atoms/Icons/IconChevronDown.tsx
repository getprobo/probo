import type { IconProps } from "./type";

export function IconChevronDown({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M3.35147 8.15147C3.8201 7.68284 4.5799 7.68284 5.04853 8.15147L12 15.1029L18.9515 8.15147C19.4201 7.68284 20.1799 7.68284 20.6485 8.15147C21.1172 8.6201 21.1172 9.3799 20.6485 9.84853L12.8485 17.6485C12.3799 18.1172 11.6201 18.1172 11.1515 17.6485L3.35147 9.84853C2.88284 9.3799 2.88284 8.6201 3.35147 8.15147Z" fill="currentColor" />
    </svg>
  );
}
