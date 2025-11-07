import { graphql } from "relay-runtime";
import {
  Button,
  Tr,
  Td,
  Table,
  Thead,
  Tbody,
  Th,
  IconChevronDown,
  Field,
  Option,
  Badge,
  IconPencil,
  IconTrashCan,
  IconArrowLink,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { TrustCenterFilesCardFragment$key, TrustCenterFilesCardFragment$data } from "./__generated__/TrustCenterFilesCardFragment.graphql";
import { useFragment } from "react-relay";
import { useMemo, useState, useCallback, useEffect } from "react";
import { sprintf, getTrustCenterVisibilityOptions } from "@probo/helpers";
import { formatDate } from "@probo/helpers";
import { IfAuthorized } from "/permissions/IfAuthorized";
import { isAuthorized } from "/permissions/permissions";
import { useParams } from "react-router";

const trustCenterFileFragment = graphql`
  fragment TrustCenterFilesCardFragment on TrustCenterFile {
    id
    name
    category
    fileUrl
    trustCenterVisibility
    createdAt
    updatedAt
  }
`;

type Mutation<Params> = (p: {
  variables: {
    input: {
      id: string;
      trustCenterVisibility: "NONE" | "PRIVATE" | "PUBLIC";
    } & Params;
  };
}) => void;

type Props<Params> = {
  files: TrustCenterFilesCardFragment$key[];
  params: Params;
  disabled?: boolean;
  onChangeVisibility: Mutation<Params>;
  onEdit: (file: { id: string; name: string; category: string }) => void;
  onDelete: (id: string) => void;
};

export function TrustCenterFilesCard<Params>(props: Props<Params>) {
  const { __ } = useTranslate();
  const [limit, setLimit] = useState<number | null>(100);
  const files = useMemo(() => {
    return limit ? props.files.slice(0, limit) : props.files;
  }, [props.files, limit]);
  const showMoreButton = limit !== null && props.files.length > limit;

  const onChangeVisibility = (fileId: string, trustCenterVisibility: "NONE" | "PRIVATE" | "PUBLIC") => {
    props.onChangeVisibility({
      variables: {
        input: {
          id: fileId,
          trustCenterVisibility,
          ...props.params,
        },
      },
    });
  };

  return (
    <div className="space-y-[10px]">
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Category")}</Th>
            <Th>{__("Upload Date")}</Th>
            <Th>{__("Visibility")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {files.length === 0 && (
            <Tr>
              <Td colSpan={5} className="text-center text-txt-secondary">
                {__("No files available")}
              </Td>
            </Tr>
          )}
          {files.map((fileFragmentRef, index) => (
            <FileRowWrapper
              key={index}
              fileFragmentRef={fileFragmentRef}
              onChangeVisibility={onChangeVisibility}
              onEdit={props.onEdit}
              onDelete={props.onDelete}
              disabled={props.disabled}
            />
          ))}
        </Tbody>
      </Table>
      {showMoreButton && (
        <Button
          variant="tertiary"
          onClick={() => setLimit(null)}
          className="mt-3 mx-auto"
          icon={IconChevronDown}
        >
          {sprintf(__("Show %s more"), props.files.length - limit)}
        </Button>
      )}
    </div>
  );
}

function FileRowWrapper(props: {
  fileFragmentRef: TrustCenterFilesCardFragment$key;
  onChangeVisibility: (fileId: string, trustCenterVisibility: "NONE" | "PRIVATE" | "PUBLIC") => void;
  onEdit: (file: { id: string; name: string; category: string }) => void;
  onDelete: (id: string) => void;
  disabled?: boolean;
}) {
  const file = useFragment(trustCenterFileFragment, props.fileFragmentRef);
  return (
    <FileRow
      file={file}
      onChangeVisibility={props.onChangeVisibility}
      onEdit={props.onEdit}
      onDelete={props.onDelete}
      disabled={props.disabled}
    />
  );
}

function FileRow(props: {
  file: TrustCenterFilesCardFragment$data;
  onChangeVisibility: (fileId: string, trustCenterVisibility: "NONE" | "PRIVATE" | "PUBLIC") => void;
  onEdit: (file: { id: string; name: string; category: string }) => void;
  onDelete: (id: string) => void;
  disabled?: boolean;
}) {
  const file = props.file;
  const { __ } = useTranslate();
  const [optimisticValue, setOptimisticValue] = useState<string | null>(null);
  const { organizationId } = useParams();

  const canUpdate = organizationId ? isAuthorized(organizationId, "TrustCenter", "update") : false;

  const handleValueChange = useCallback((value: string | {}) => {
    const stringValue = typeof value === 'string' ? value : '';
    const typedValue = stringValue as "NONE" | "PRIVATE" | "PUBLIC";
    setOptimisticValue(typedValue);
    props.onChangeVisibility(file.id, typedValue);
  }, [file.id, props.onChangeVisibility]);

  useEffect(() => {
    if (optimisticValue && file.trustCenterVisibility === optimisticValue) {
      setOptimisticValue(null);
    }
  }, [file.trustCenterVisibility, optimisticValue]);

  const currentValue = optimisticValue || file.trustCenterVisibility;

  const visibilityOptions = getTrustCenterVisibilityOptions(__);

  return (
    <Tr>
      <Td>
        <div className="flex gap-4 items-center">
          {file.name}
        </div>
      </Td>
      <Td>{file.category}</Td>
      <Td>{formatDate(file.createdAt)}</Td>
      <Td noLink width={130} className="pr-0">
        <Field
          type="select"
          value={currentValue}
          onValueChange={handleValueChange}
          disabled={props.disabled || !canUpdate}
          className="w-[105px]"
        >
          {visibilityOptions.map((option) => (
            <Option key={option.value} value={option.value}>
              <div className="flex items-center justify-between w-full">
                <Badge variant={option.variant}>
                  {option.label}
                </Badge>
              </div>
            </Option>
          ))}
        </Field>
      </Td>
      <Td noLink width={120}>
        <div className="flex gap-2">
          <Button
            variant="secondary"
            icon={IconArrowLink}
            onClick={() => window.open(file.fileUrl, '_blank', 'noopener,noreferrer')}
            title={__("Download")}
          />
          <IfAuthorized entity="TrustCenterFile" action="update">
            <Button
              variant="secondary"
              icon={IconPencil}
              onClick={() => props.onEdit({ id: file.id, name: file.name, category: file.category })}
              disabled={props.disabled}
              title={__("Edit")}
            />
          </IfAuthorized>
          <IfAuthorized entity="TrustCenterFile" action="delete">
            <Button
              variant="danger"
              icon={IconTrashCan}
              onClick={() => props.onDelete(file.id)}
              disabled={props.disabled}
              title={__("Delete")}
            />
          </IfAuthorized>
        </div>
      </Td>
    </Tr>
  );
}
