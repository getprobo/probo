import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Select,
  Option,
  Textarea,
  useDialogRef,
  Spinner,
  Badge,
  IconCheckmark1,
  IconTrashCan,
  IconPencil,
  Input,
  IconMagnifyingGlass,
} from "@probo/ui";
import { forwardRef, useImperativeHandle, useState, Suspense, useMemo } from "react";
import { useLazyLoadQuery } from "react-relay";
import { graphql } from "react-relay";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";

const linkControlQuery = graphql`
  query LinkControlDialogQuery($stateOfApplicabilityId: ID!) {
    node(id: $stateOfApplicabilityId) {
      ... on StateOfApplicability {
        id
        availableControls {
          controlId
          sectionTitle
          name
          frameworkId
          frameworkName
          organizationId
          stateOfApplicabilityId
          state
          exclusionJustification
        }
      }
    }
  }
`;

const linkControlMutation = graphql`
  mutation LinkControlDialogLinkMutation($input: LinkStateOfApplicabilityControlInput!) {
    linkStateOfApplicabilityControl(input: $input) {
      stateOfApplicabilityControl {
        stateOfApplicabilityId
        controlId
        state
        exclusionJustification
      }
    }
  }
`;

const unlinkControlMutation = graphql`
  mutation LinkControlDialogUnlinkMutation($input: UnlinkStateOfApplicabilityControlInput!) {
    unlinkStateOfApplicabilityControl(input: $input) {
      deletedControlId
    }
  }
`;

export type LinkControlDialogRef = {
  open: (stateOfApplicabilityId: string, onSuccess?: () => void) => void;
};

const stateLabels: Record<string, string> = {
  IMPLEMENTED: "Implemented",
  NOT_IMPLEMENTED: "Not Implemented",
  EXCLUDED: "Excluded",
};

type Control = {
  controlId: string;
  sectionTitle: string;
  name: string;
  frameworkId: string;
  frameworkName: string;
  organizationId: string;
  stateOfApplicabilityId: string | null;
  state: string | null;
  exclusionJustification: string | null;
};

function ControlRow({
  control,
  stateOfApplicabilityId,
  onLink,
  onUnlink,
  isLinked,
}: {
  control: Control;
  stateOfApplicabilityId: string;
  onLink: (controlId: string) => void;
  onUnlink: (controlId: string) => void;
  isLinked: boolean;
}) {
  const { __ } = useTranslate();
  const [isEditing, setIsEditing] = useState(false);
  const [state, setState] = useState<"EXCLUDED" | "IMPLEMENTED" | "NOT_IMPLEMENTED">(
    (control.state as "EXCLUDED" | "IMPLEMENTED" | "NOT_IMPLEMENTED") || "IMPLEMENTED"
  );
  const [exclusionJustification, setExclusionJustification] = useState(
    control.exclusionJustification || ""
  );

  const [linkMutate, isLinking] = useMutationWithToasts(linkControlMutation, {
    successMessage: __("Control added successfully."),
    errorMessage: __("Failed to add control"),
  });

  const [unlinkMutate, isUnlinking] = useMutationWithToasts(unlinkControlMutation, {
    successMessage: __("Control removed successfully."),
    errorMessage: __("Failed to remove control"),
  });

  const handleSave = () => {
    linkMutate({
      variables: {
        input: {
          stateOfApplicabilityId,
          controlId: control.controlId,
          state,
          exclusionJustification: state === "EXCLUDED" ? exclusionJustification || null : null,
        },
      },
      onSuccess: () => {
        setIsEditing(false);
        onLink(control.controlId);
      },
      updater: (store) => {
        const stateOfApplicability = store.get(stateOfApplicabilityId);
        if (stateOfApplicability) {
          stateOfApplicability.invalidateRecord();
        }
      },
    });
  };

  const handleUnlink = () => {
    unlinkMutate({
      variables: {
        input: {
          stateOfApplicabilityId,
          controlId: control.controlId,
        },
      },
      onSuccess: () => {
        onUnlink(control.controlId);
      },
      updater: (store) => {
        const stateOfApplicability = store.get(stateOfApplicabilityId);
        if (stateOfApplicability) {
          stateOfApplicability.invalidateRecord();
        }
      },
    });
  };

  if (isEditing) {
    return (
      <div className="p-4 border-b border-border-low space-y-4">
        <div>
          <div className="text-sm font-medium text-txt-primary mb-1">
            {control.sectionTitle}: {control.name}
          </div>
        </div>

        <div className="flex items-center gap-2">
          <label className="text-sm font-medium text-txt-primary whitespace-nowrap">
            {__("State")}:
          </label>
          <Select
            variant="editor"
            value={state}
            onValueChange={(value) => setState(value as "EXCLUDED" | "IMPLEMENTED" | "NOT_IMPLEMENTED")}
          >
            <Option value="IMPLEMENTED">{__("Implemented")}</Option>
            <Option value="NOT_IMPLEMENTED">{__("Not Implemented")}</Option>
            <Option value="EXCLUDED">{__("Excluded")}</Option>
          </Select>
        </div>

        {state === "EXCLUDED" && (
          <Field label={__("Exclusion Justification")}>
            <Textarea
              value={exclusionJustification}
              onChange={(e) => setExclusionJustification(e.target.value)}
              placeholder={__("Reason for exclusion")}
              autogrow
            />
          </Field>
        )}

        <div className="flex gap-2 justify-end">
          <Button
            variant="secondary"
            onClick={() => {
              setIsEditing(false);
              setState((control.state as "EXCLUDED" | "IMPLEMENTED" | "NOT_IMPLEMENTED") || "IMPLEMENTED");
              setExclusionJustification(control.exclusionJustification || "");
            }}
          >
            {__("Cancel")}
          </Button>
          <Button onClick={handleSave} disabled={isLinking}>
            {isLinked ? __("Update") : __("Add")}
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="p-4 border-b border-border-low">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <Badge size="md">{control.sectionTitle}</Badge>
            <span className="text-sm font-medium text-txt-primary">{control.name}</span>
            {isLinked && (
              <IconCheckmark1 size={16} className="text-success" />
            )}
          </div>
          {isLinked && control.state && (
            <div className="mt-2 text-sm">
              <span className="text-txt-secondary">{__("State")}: </span>
              <span>{__(stateLabels[control.state] || control.state)}</span>
              {control.exclusionJustification && (
                <>
                  <span className="text-txt-secondary ml-4">{__("Justification")}: </span>
                  <span>{control.exclusionJustification}</span>
                </>
              )}
            </div>
          )}
        </div>
        <div className="flex gap-2">
          {isLinked ? (
            <>
              <Button
                variant="secondary"
                icon={IconPencil}
                onClick={() => setIsEditing(true)}
                disabled={isUnlinking}
              >
                {__("Edit")}
              </Button>
              <Button
                variant="secondary"
                icon={IconTrashCan}
                onClick={handleUnlink}
                disabled={isUnlinking}
              >
                {__("Remove")}
              </Button>
            </>
          ) : (
            <Button
              variant="secondary"
              icon={IconPencil}
              onClick={() => setIsEditing(true)}
            >
              {__("Add")}
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}

function LinkControlDialogContent({
  stateOfApplicabilityId,
  onSuccess,
}: {
  stateOfApplicabilityId: string;
  onSuccess: () => void;
}) {
  const { __ } = useTranslate();
  const [search, setSearch] = useState("");
  const [refetchKey, setRefetchKey] = useState(0);
  const data = useLazyLoadQuery(
    linkControlQuery,
    { stateOfApplicabilityId },
    { fetchPolicy: "network-only", fetchKey: refetchKey }
  ) as {
    node: {
      availableControls?: Control[];
    } | null;
  };

  const controls = data.node?.availableControls || [];

  const filteredControls = useMemo(() => {
    if (!search) return controls;
    const lowerSearch = search.toLowerCase();
    return controls.filter(
      (c) =>
        c.name.toLowerCase().includes(lowerSearch) ||
        c.sectionTitle.toLowerCase().includes(lowerSearch) ||
        c.frameworkName.toLowerCase().includes(lowerSearch)
    );
  }, [controls, search]);

  const groupedControls = useMemo(() => {
    const groups: Record<string, Record<string, Control[]>> = {};
    filteredControls.forEach((control) => {
      if (!groups[control.frameworkName]) {
        groups[control.frameworkName] = {};
      }
      if (!groups[control.frameworkName][control.sectionTitle]) {
        groups[control.frameworkName][control.sectionTitle] = [];
      }
      groups[control.frameworkName][control.sectionTitle].push(control);
    });
    return groups;
  }, [filteredControls]);

  const handleLink = () => {
    setRefetchKey((k) => k + 1);
  };

  const handleUnlink = () => {
    setRefetchKey((k) => k + 1);
  };

  return (
    <>
      <DialogContent className="p-0">
        <div className="sticky top-0 bg-level-2 p-4 border-b border-border-low z-10">
          <Input
            icon={IconMagnifyingGlass}
            placeholder={__("Search controls...")}
            onValueChange={setSearch}
          />
        </div>
        <div className="max-h-[60vh] overflow-y-auto">
          {filteredControls.length === 0 ? (
            <div className="p-8 text-center text-txt-secondary">
              {__("No controls found")}
            </div>
          ) : (
            Object.entries(groupedControls).map(([frameworkName, sections]) => (
              <div key={frameworkName}>
                <div className="sticky top-0 bg-level-1 px-4 py-2 border-b border-border-low z-10">
                  <h3 className="text-sm font-semibold text-txt-primary">{frameworkName}</h3>
                </div>
                {Object.entries(sections).map(([sectionTitle, sectionControls]) => (
                  <div key={`${frameworkName}-${sectionTitle}`}>
                    {sectionControls.map((control) => (
                      <ControlRow
                        key={control.controlId}
                        control={control}
                        stateOfApplicabilityId={stateOfApplicabilityId}
                        onLink={handleLink}
                        onUnlink={handleUnlink}
                        isLinked={control.stateOfApplicabilityId !== null}
                      />
                    ))}
                  </div>
                ))}
              </div>
            ))
          )}
        </div>
      </DialogContent>
      <DialogFooter>
        <Button variant="secondary" onClick={() => onSuccess()}>
          {__("Close")}
        </Button>
      </DialogFooter>
    </>
  );
}

export const LinkControlDialog = forwardRef<LinkControlDialogRef>((_props, ref) => {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const [stateOfApplicabilityId, setStateOfApplicabilityId] = useState<string | null>(null);
  const [onSuccessCallback, setOnSuccessCallback] = useState<(() => void) | undefined>(undefined);

  useImperativeHandle(ref, () => ({
    open: (soaId: string, callback?: () => void) => {
      setStateOfApplicabilityId(soaId);
      setOnSuccessCallback(() => callback);
      dialogRef.current?.open();
    },
  }), []);

  const handleClose = () => {
    dialogRef.current?.close();
    onSuccessCallback?.();
    setStateOfApplicabilityId(null);
    setOnSuccessCallback(undefined);
  };

  return (
    <Dialog
      ref={dialogRef}
      className="max-w-3xl"
      title={
        <Breadcrumb
          items={[__("States of Applicability"), __("Manage Controls")]}
        />
      }
    >
      {stateOfApplicabilityId ? (
        <Suspense
          fallback={
            <DialogContent padded className="flex items-center justify-center py-8">
              <Spinner />
            </DialogContent>
          }
        >
          <LinkControlDialogContent
            stateOfApplicabilityId={stateOfApplicabilityId}
            onSuccess={handleClose}
          />
        </Suspense>
      ) : null}
    </Dialog>
  );
});
