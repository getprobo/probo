import type { IconProps } from "./type";

export function IconCalendar2({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M8 2C8.55228 2 9 2.44772 9 3V4H15V3C15 2.44772 15.4477 2 16 2C16.5523 2 17 2.44772 17 3V4C19.2092 4 21 5.79088 21 8V17C21 19.2092 19.2092 21 17 21H7C4.79088 21 3 19.2092 3 17V8C3 5.79087 4.79087 4 7 4V3C7 2.44772 7.44772 2 8 2ZM7 6C5.89543 6 5 6.89543 5 8V9H19V8C19 6.89542 18.1046 6 17 6H7ZM19 11H5V17C5 18.1046 5.89542 19 7 19H17C18.1046 19 19 18.1046 19 17V11Z" fill="currentColor" />
    </svg>
  );
}
