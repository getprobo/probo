import type { IconProps } from "./type";

export function IconChevronRight({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M8.15147 3.35147C8.6201 2.88284 9.37989 2.88284 9.84852 3.35147L17.6485 11.1515C18.1172 11.6201 18.1172 12.3799 17.6485 12.8485L9.84852 20.6485C9.37989 21.1172 8.6201 21.1172 8.15147 20.6485C7.68284 20.1799 7.68284 19.4201 8.15147 18.9515L15.1029 12L8.15147 5.04853C7.68284 4.5799 7.68284 3.8201 8.15147 3.35147Z" fill="currentColor" />
    </svg>
  );
}
