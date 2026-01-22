import type { IconProps } from "./type";

export function IconCalendar1({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path d="M8 2C8.55228 2 9 2.44772 9 3V4H15V3C15 2.44772 15.4477 2 16 2C16.5523 2 17 2.44772 17 3V4C19.2091 4 21 5.79086 21 8V9H3V8C3 5.79086 4.79086 4 7 4V3C7 2.44772 7.44772 2 8 2Z" fill="currentColor" />
      <path d="M3 11V17C3 19.2091 4.79086 21 7 21H17C19.2091 21 21 19.2091 21 17V11H3Z" fill="currentColor" />
    </svg>
  );
}
