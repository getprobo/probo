import { availableFrameworks } from "@probo/helpers";
import { Avatar } from "@probo/ui";

const availableLogos = new Map(
  availableFrameworks.map((framework) => [framework.name, framework.logo])
);

export function FrameworkLogo({ name }: { name: string }) {
  const logo = availableLogos.get(name);
  return logo ? (
    <img src={logo} alt="" className="size-12" />
  ) : (
    <Avatar name={name} size="l" className="size-12" />
  );
}
