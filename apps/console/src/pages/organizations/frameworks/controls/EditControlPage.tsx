import { Loader2 } from "lucide-react";
import EditControlView from "./EditControlView";

export function EditControlViewSkeleton() {
  return (
    <div className="flex items-center justify-center h-full">
      <Loader2 className="w-8 h-8 animate-spin text-info" />
    </div>
  );
}

export function EditControlPage() {
  return <EditControlView />;
}
