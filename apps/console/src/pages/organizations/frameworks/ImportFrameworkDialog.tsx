import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Plus, Loader2 } from "lucide-react";
import { useNavigate, useParams } from "react-router";
import { graphql, useMutation } from "react-relay";
import { useToast } from "@/hooks/use-toast";
import { type ImportFrameworkDialogImportFrameworkMutation } from "./__generated__/ImportFrameworkDialogImportFrameworkMutation.graphql";
import { useState } from "react";

const AVAILABLE_FRAMEWORKS = [
  { id: "ISO27001-2022", name: "ISO/IEC 27001:2022" },
  { id: "SOC2", name: "SOC 2" },
];

const importFrameworkMutation = graphql`
  mutation ImportFrameworkDialogImportFrameworkMutation(
    $input: ImportFrameworkInput!
    $connections: [ID!]!
  ) {
    importFramework(input: $input) {
      frameworkEdge @prependEdge(connections: $connections) {
        node {
          id
          name
        }
      }
    }
  }
`;

export function FrameworkImportDropdown() {
  const { organizationId } = useParams();
  const navigate = useNavigate();
  const { toast } = useToast();
  const [loadingFramework, setLoadingFramework] = useState<string | null>(null);
  const [commit, isInFlight] =
    useMutation<ImportFrameworkDialogImportFrameworkMutation>(
      importFrameworkMutation,
    );

  const handleImport = async (frameworkId: string) => {
    /*
    setLoadingFramework(frameworkId);
    try {
      const frameworkTemplate = await import(
        `./../../../../../data/frameworks/${frameworkId}.json`,
        { with: { type: "json" } }
      );

      const file = new File(
        [JSON.stringify(frameworkTemplate.default)],
        `${frameworkId}.json`,
        {
          type: "application/json",
        },
      );

      const connectionId = ConnectionHandler.getConnectionID(
        organizationId!,
        "FrameworkListView_frameworks",
      );

      commit({
        variables: {
          input: {
            organizationId: organizationId!,
            file: null,
          },
          connections: [connectionId],
        },
        uploadables: {
          "input.file": file,
        },
        onCompleted(data, errors) {
          if (errors) {
            toast({
              title: "Error",
              description: errors[0]?.message || "Failed to import framework",
              variant: "destructive",
            });
            return;
          }

          toast({
            title: "Success",
            description: "Framework imported successfully",
          });
        },
        onError(error) {
          toast({
            title: "Error",
            description: error.message || "Failed to import framework",
            variant: "destructive",
          });
        },
      });
    } catch (error) {
      toast({
        title: "Error",
        description:
          error instanceof Error
            ? error.message
            : "Failed to load framework template",
        variant: "destructive",
      });
    } finally {
      setLoadingFramework(null);
    }
    */
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button disabled={isInFlight || loadingFramework !== null}>
          {isInFlight || loadingFramework !== null ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              {isInFlight ? "Importing..." : "Loading..."}
            </>
          ) : (
            <>
              <Plus className="mr-2 h-4 w-4" />
              Create
            </>
          )}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem
          onClick={() =>
            navigate(`/organizations/${organizationId}/frameworks/new`)
          }
        >
          Custom Framework
        </DropdownMenuItem>
        {AVAILABLE_FRAMEWORKS.map((framework) => (
          <DropdownMenuItem
            key={framework.id}
            onClick={() => handleImport(framework.id)}
            disabled={loadingFramework === framework.id}
          >
            {loadingFramework === framework.id ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Loading...
              </>
            ) : (
              framework.name
            )}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
