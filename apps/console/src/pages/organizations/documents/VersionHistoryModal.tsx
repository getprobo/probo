import { Dialog, DialogContent } from "@/components/ui/dialog";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { format } from "date-fns";
import { Button } from "@/components/ui/button";
import { useState } from "react";
import ReactMarkdown from "react-markdown";
import { graphql, useFragment } from "react-relay";
import type { VersionHistoryModal_documentVersions$key } from "./__generated__/VersionHistoryModal_documentVersions.graphql";
import remarkGfm from "remark-gfm";
import rehypeRaw from "rehype-raw";

export const documentVersionsFragment = graphql`
  fragment VersionHistoryModal_documentVersions on Document {
    title
    owner {
      fullName
    }
    versionHistory: versions(first: 20) {
      edges {
        node {
          id
          version
          status
          content
          changelog
          publishedAt
          updatedAt
          publishedBy {
            fullName
          }
        }
      }
    }
  }
`;

interface VersionHistoryModalProps {
  isOpen: boolean;
  onClose: () => void;
  documentRef: VersionHistoryModal_documentVersions$key;
  onRestoreVersion?: (versionNumber: number) => void;
}

export function VersionHistoryModal({
  isOpen,
  onClose,
  documentRef,
  onRestoreVersion,
}: VersionHistoryModalProps) {
  const data = useFragment<VersionHistoryModal_documentVersions$key>(
    documentVersionsFragment,
    documentRef,
  );

  // Safely access and extract version nodes
  const versionNodes =
    data?.versionHistory?.edges?.map((edge) => edge.node) || [];

  // Sort versions by version number (descending)
  const versions = [...versionNodes].sort((a, b) => b.version - a.version);

  const latestVersion = versions.length > 0 ? versions[0].version : 0;
  const [selectedVersion, setSelectedVersion] = useState<number>(latestVersion);

  // Format time helper for version history
  const formatDateTime = (dateString?: string | null) => {
    if (!dateString) return "N/A";
    const date = new Date(dateString);
    return format(date, "h:mm a • MMM d, yyyy");
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-[1080px] h-[744px] p-0 overflow-hidden flex flex-col">
        <div className="flex flex-1 overflow-hidden">
          {/* Version History Sidebar */}
          <div className="w-[316px] border-r border-solid-b h-full overflow-y-auto">
            <h2 className="text-lg font-semibold p-6 pb-4">Version history</h2>
            <div className="overflow-y-auto">
              {versions.map((version) => (
                <div
                  key={version.id}
                  className={`flex items-center gap-3 p-6 py-4 cursor-pointer ${selectedVersion === version.version ? "bg-slate-50" : ""}`}
                  onClick={() => setSelectedVersion(version.version)}
                >
                  <Avatar className="h-10 w-10">
                    <AvatarImage
                      src=""
                      alt={
                        version.publishedBy?.fullName ||
                        data.owner?.fullName ||
                        ""
                      }
                    />
                    <AvatarFallback>
                      {(
                        version.publishedBy?.fullName ||
                        data.owner?.fullName ||
                        ""
                      ).charAt(0)}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <span className="font-medium">
                        Version {version.version}
                      </span>
                      {version.status === "PUBLISHED" ? (
                        <span className="text-xs px-2 py-0.5 rounded-full bg-green-100 text-green-800">
                          Published
                        </span>
                      ) : (
                        <span className="text-xs px-2 py-0.5 rounded-full bg-yellow-100 text-yellow-800">
                          Draft
                        </span>
                      )}
                    </div>
                    <div className="text-sm text-tertiary">
                      {version.publishedBy?.fullName ||
                        data.owner?.fullName ||
                        "Unknown"}{" "}
                      •{" "}
                      {formatDateTime(version.publishedAt || version.updatedAt)}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Content Area */}
          <div className="flex-1 p-6 relative overflow-y-auto">
            <h2 className="text-2xl font-semibold mb-6">{data.title}</h2>

            <div className="prose prose-olive max-w-none">
              {selectedVersion && (
                <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeRaw]}>
                  {versions.find((v) => v.version === selectedVersion)
                    ?.content || "No content available"}
                </ReactMarkdown>
              )}
            </div>
          </div>
        </div>

        <div className="flex justify-end p-4 border-t border-solid-b mt-auto">
          <Button variant="outline" onClick={onClose} className="mr-2">
            Cancel
          </Button>

          {onRestoreVersion && (
            <Button
              variant="default"
              onClick={() => onRestoreVersion(selectedVersion)}
              disabled={selectedVersion === latestVersion}
            >
              Restore version
            </Button>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
