export function Divider({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative my-6 w-full">
      <div className="border-t border-border-mid" />
      <span className="px-4 text-xs uppercase text-txt-secondary bg-level-0 absolute top-0 left-1/2 -translate-1/2">
        {children}
      </span>
    </div>
  );
}
