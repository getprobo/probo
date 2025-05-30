import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { format } from "date-fns";

interface DocumentVersionSignature {
  id: string;
  state: "REQUESTED" | "SIGNED";
  signedBy: {
    fullName: string;
  };
  signedAt?: string;
  requestedAt: string;
  requestedBy: {
    fullName: string;
  };
}

interface SignaturesListProps {
  signatures: DocumentVersionSignature[];
}

export function SignaturesList({ signatures }: SignaturesListProps) {
  const formatDateTime = (dateString: string) => {
    const date = new Date(dateString);
    return format(date, "h:mm a • MMM d, yyyy");
  };

  return (
    <div className="space-y-4">
      {signatures.map((signature) => (
        <div
          key={signature.id}
          className="bg-white rounded-lg border border-solid-b shadow-sm p-6"
        >
          <div className="flex items-center gap-3 mb-4">
            <Avatar className="h-10 w-10">
              <AvatarImage src="" alt={signature.signedBy.fullName} />
              <AvatarFallback>
                {signature.signedBy.fullName.charAt(0)}
              </AvatarFallback>
            </Avatar>
            <div>
              <div className="font-medium">{signature.signedBy.fullName}</div>
              <div className="text-sm text-tertiary">
                {signature.state === "SIGNED" && signature.signedAt
                  ? formatDateTime(signature.signedAt)
                  : formatDateTime(signature.requestedAt)}
              </div>
            </div>
          </div>

          <div className="space-y-2">
            <div className="text-sm">
              <span className="text-muted-foreground">Status:</span>{" "}
              <span
                className={`px-2 py-0.5 rounded-full ${
                  signature.state === "SIGNED"
                    ? "bg-green-100 text-green-800"
                    : "bg-yellow-100 text-yellow-800"
                }`}
              >
                {signature.state}
              </span>
            </div>
            <div className="text-sm">
              <span className="text-muted-foreground">Requested by:</span>{" "}
              {signature.requestedBy.fullName}
            </div>
            {signature.state === "SIGNED" && signature.signedAt && (
              <div className="text-sm">
                <span className="text-muted-foreground">Signed at:</span>{" "}
                {formatDateTime(signature.signedAt)}
              </div>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
