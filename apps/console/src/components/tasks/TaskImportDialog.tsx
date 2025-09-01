import {
  Button,
  Dialog,
  DialogContent,
  IconUpload,
  IconPageTextLine,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  Badge,
  Card,
  useDialogRef,
} from "@probo/ui";
import { useState, useCallback } from "react";
import { graphql, useMutation } from "react-relay";
import { useTranslate } from "@probo/i18n";
import { promisifyMutation } from "@probo/helpers";
import { useOrganizationId } from "/hooks/useOrganizationId";
import type { TaskImportDialogMutation$data } from "./__generated__/TaskImportDialogMutation.graphql";

type Props = {
  children: React.ReactNode;
};

const importTasksMutation = graphql`
  mutation TaskImportDialogMutation(
    $input: ImportTasksInput!
  ) {
    importTasks(input: $input) {
      importResults {
        rowNumber
        success
        error
        task {
          id
          name
          state
          description
          ...TaskFormDialogFragment
          measure {
            id
            name
          }
          assignedTo {
            id
            fullName
          }
        }
      }
      successCount
      errorCount
    }
  }
`;

type ImportResult = {
  rowNumber: number;
  success: boolean;
  error?: string | null;
  task?: {
    id: string;
    name: string;
  } | null;
};

export default function TaskImportDialog({ children }: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const dialogRef = useDialogRef();
  const [file, setFile] = useState<File | null>(null);
  const [importResults, setImportResults] = useState<ImportResult[] | null>(null);
  const [isImporting, setIsImporting] = useState(false);
  
  const [importTasks] = useMutation(importTasksMutation);

  const handleFileSelect = useCallback((files: File[]) => {
    if (files && files.length > 0) {
      const selectedFile = files[0];
      if (selectedFile.type === 'text/csv' || selectedFile.name.endsWith('.csv')) {
        setFile(selectedFile);
        setImportResults(null); // Clear previous results
      } else {
        alert(__('Please select a CSV file'));
      }
    }
  }, [__]);

  const handleImport = useCallback(async () => {
    if (!file) return;

    setIsImporting(true);
    try {
      const result = await promisifyMutation(importTasks)({
        variables: {
          input: {
            organizationId,
            file: null,
          },
        },
        uploadables: {
          "input.file": file,
        },
      }) as TaskImportDialogMutation$data;

      if (result?.importTasks) {
        setImportResults([...result.importTasks.importResults] as ImportResult[]);
        // Refresh the page to show newly imported tasks
        // In a production app, you'd want to refetch the query instead
        if (result.importTasks.successCount > 0) {
          setTimeout(() => window.location.reload(), 2000);
        }
      }
    } catch (error) {
      console.error('Import failed:', error);
      alert(__('Import failed. Please try again.'));
    } finally {
      setIsImporting(false);
    }
  }, [file, organizationId, importTasks, __]);

  const handleClose = useCallback(() => {
    dialogRef.current?.close();
    setFile(null);
    setImportResults(null);
  }, []);

  const renderResults = () => {
    if (!importResults) return null;

    const successCount = importResults.filter(r => r.success).length;
    const errorCount = importResults.filter(r => !r.success).length;

    return (
      <div className="space-y-4">
        <Card className={`p-4 ${errorCount > 0 ? 'border-red-200 bg-red-50' : 'border-green-200 bg-green-50'}`}>
          <h3 className={`font-medium ${errorCount > 0 ? 'text-red-800' : 'text-green-800'}`}>
            {__("Import Complete")}
          </h3>
          <p className={`text-sm ${errorCount > 0 ? 'text-red-700' : 'text-green-700'}`}>
            {__("Successfully imported")} {successCount} {__("tasks.")} {errorCount} {__("errors.")}
          </p>
        </Card>

        {errorCount > 0 && (
          <div>
            <h3 className="font-medium mb-2">{__("Import Details")}</h3>
            <div className="max-h-64 overflow-y-auto border rounded">
              <Table>
                <Thead>
                  <Tr>
                    <Th>{__("Row")}</Th>
                    <Th>{__("Status")}</Th>
                    <Th>{__("Task Name")}</Th>
                    <Th>{__("Error")}</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {importResults.map((result) => (
                    <Tr key={result.rowNumber}>
                      <Td>{result.rowNumber}</Td>
                      <Td>
                        <Badge variant={result.success ? "success" : "danger"}>
                          {result.success ? __("Success") : __("Error")}
                        </Badge>
                      </Td>
                      <Td>{result.task?.name || "-"}</Td>
                      <Td className="text-red-600">{result.error || "-"}</Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            </div>
          </div>
        )}
      </div>
    );
  };

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={__("Import Tasks from CSV")}
    >
      <DialogContent className="max-w-2xl">
        <div className="space-y-6 p-6">
          <div>
            <h3 className="font-medium mb-2">{__("CSV Format")}</h3>
            <div className="text-sm text-gray-600 space-y-1">
              <p>{__("Required column: name")}</p>
              <p>{__("Optional columns: description, measureId, assignedTo, deadline (YYYY-MM-DD), timeEstimate (minutes)")}</p>
            </div>
          </div>

          {!importResults && (
            <div className="border-2 border-dashed border-gray-300 rounded-lg p-8 text-center hover:border-gray-400">
              <div className="space-y-2">
                <IconUpload className="mx-auto h-12 w-12 text-gray-400" />
                <div>
                  <p className="text-sm font-medium">
                    {file ? (
                      <span className="flex items-center justify-center gap-2">
                        <IconPageTextLine size={16} />
                        {file.name}
                      </span>
                    ) : (
                      __("Drop CSV file here or click to browse")
                    )}
                  </p>
                  <input
                    type="file"
                    accept=".csv,text/csv"
                    onChange={(e) => handleFileSelect(Array.from(e.target.files || []))}
                    className="hidden"
                    id="csv-file-input"
                  />
                  <label
                    htmlFor="csv-file-input"
                    className="cursor-pointer text-blue-600 hover:underline"
                  >
                    {__("Select file")}
                  </label>
                </div>
              </div>
            </div>
          )}

          {renderResults()}

          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={handleClose}>
              {importResults ? __("Close") : __("Cancel")}
            </Button>
            {file && !importResults && (
              <Button 
                onClick={handleImport} 
                disabled={isImporting}
                icon={IconUpload}
              >
                {isImporting ? __("Importing...") : __("Import Tasks")}
              </Button>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}