import clsx from "clsx";

type Props = {
  className?: string;
  withPicto?: boolean;
};

export function Logo({ className, withPicto }: Props) {
  const src = "/logos/govrly.png";
  const alt = "Govrly";

  if (withPicto) {
    return (
      <img
        src={src}
        alt={alt}
        className={clsx(className, "aspect-auto h-auto")}
      />
    );
  }

  return <img src={src} alt={alt} className={clsx(className, "h-auto")} />;
}
