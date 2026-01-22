import type { IconProps } from "./type";

export function IconCheckmark1({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M19.3209 4.24472C20.0143 4.69807 20.2088 5.62768 19.7555 6.32105L11.2555 19.321C10.9972 19.7161 10.5681 19.9665 10.0971 19.997C9.62616 20.0276 9.16828 19.8347 8.86114 19.4764L4.36114 14.2264C3.82201 13.5974 3.89485 12.6504 4.52384 12.1113C5.15283 11.5722 6.09978 11.645 6.63891 12.274L9.83828 16.0066L17.2446 4.6793C17.6979 3.98593 18.6275 3.79136 19.3209 4.24472Z" fill="currentColor" />
    </svg>
  );
}
