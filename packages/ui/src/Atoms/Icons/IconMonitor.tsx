import type { IconProps } from "./type";

export function IconMonitor({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M1.25 6C1.25 4.48122 2.48122 3.25 4 3.25H20C21.5188 3.25 22.75 4.48122 22.75 6V15C22.75 16.5188 21.5188 17.75 20 17.75H12.75V19.25H16C16.4142 19.25 16.75 19.5858 16.75 20C16.75 20.4142 16.4142 20.75 16 20.75H8C7.58579 20.75 7.25 20.4142 7.25 20C7.25 19.5858 7.58579 19.25 8 19.25H11.25V17.75H4C2.48122 17.75 1.25 16.5188 1.25 15V6ZM4 4.75C3.30964 4.75 2.75 5.30964 2.75 6V15C2.75 15.6904 3.30964 16.25 4 16.25H20C20.6904 16.25 21.25 15.6904 21.25 15V6C21.25 5.30964 20.6904 4.75 20 4.75H4Z" fill="currentColor" />
    </svg>
  );
}
