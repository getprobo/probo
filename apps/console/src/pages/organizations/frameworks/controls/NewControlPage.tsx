import { Loader2 } from "lucide-react";
import NewControlView from "./NewControlView";

export function NewControlViewSkeleton() {
  return (
    <div className="flex items-center justify-center h-full">
      <Loader2 className="w-8 h-8 animate-spin text-info" />
    </div>
  );
}

export function NewControlPage() {
  return <NewControlView />;
}
