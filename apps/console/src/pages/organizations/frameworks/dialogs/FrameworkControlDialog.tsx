import {
    Breadcrumb,
    Button,
    Dialog,
    DialogContent,
    DialogFooter,
    Input,
    Textarea,
    useDialogRef,
    Checkbox,
} from "@probo/ui";
import type { ReactNode } from "react";
import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";
import { useFragment } from "react-relay";
import type { FrameworkControlDialogFragment$key } from "/__generated__/core/FrameworkControlDialogFragment.graphql";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { z } from "zod";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { useEffect, useMemo } from "react";

type Props = {
    children: ReactNode;
    control?: FrameworkControlDialogFragment$key;
    frameworkId: string;
    connectionId?: string;
};

const controlFragment = graphql`
    fragment FrameworkControlDialogFragment on Control {
        id
        name
        description
        sectionTitle
        status
        exclusionJustification
        bestPractice
    }
`;

const createMutation = graphql`
    mutation FrameworkControlDialogCreateMutation(
        $input: CreateControlInput!
        $connections: [ID!]!
    ) {
        createControl(input: $input) {
            controlEdge @prependEdge(connections: $connections) {
                node {
                    ...FrameworkControlDialogFragment
                }
            }
        }
    }
`;

const updateMutation = graphql`
    mutation FrameworkControlDialogUpdateMutation($input: UpdateControlInput!) {
        updateControl(input: $input) {
            control {
                ...FrameworkControlDialogFragment
            }
        }
    }
`;

const schema = z
    .object({
        name: z.string(),
        description: z.string().optional().nullable(),
        sectionTitle: z.string(),
        status: z.enum(["INCLUDED", "EXCLUDED"]),
        exclusionJustification: z.string().optional(),
        bestPractice: z.boolean(),
    })
    .refine(
        (data) => {
            if (data.status === "EXCLUDED") {
                return (
                    data.exclusionJustification &&
                    data.exclusionJustification.trim().length > 0
                );
            }
            return true;
        },
        {
            message:
                "Exclusion justification is required when status is excluded",
            path: ["exclusionJustification"],
        },
    );

export function FrameworkControlDialog(props: Props) {
    const { __ } = useTranslate();
    const frameworkControl = useFragment(controlFragment, props.control);
    const dialogRef = useDialogRef();
    const [mutate, isMutating] = useMutationWithToasts(
        props.control ? updateMutation : createMutation,
        {
            successMessage: __(
                `Control ${props.control ? "updated" : "created"} successfully.`,
            ),
            errorMessage: __(
                `Failed to ${props.control ? "update" : "create"} control`,
            ),
        },
    );

    const defaultValues = useMemo(
        () => ({
            name: frameworkControl?.name ?? "",
            description: frameworkControl?.description ?? "",
            sectionTitle: frameworkControl?.sectionTitle ?? "",
            status: frameworkControl?.status ?? "INCLUDED",
            exclusionJustification:
                frameworkControl?.exclusionJustification ?? "",
            bestPractice: frameworkControl?.bestPractice ?? true,
        }),
        [frameworkControl],
    );

    const { handleSubmit, register, reset, watch, setValue } =
        useFormWithSchema(schema, {
            defaultValues,
        });

    useEffect(() => {
        reset(defaultValues);
    }, [defaultValues, reset]);

    const bestPracticeValue = watch("bestPractice");

    const onSubmit = handleSubmit(async (data) => {
        if (frameworkControl) {
            // Update the control
            await mutate({
                variables: {
                    input: {
                        id: frameworkControl.id,
                        name: data.name,
                        description: data.description || null,
                        sectionTitle: data.sectionTitle,
                        bestPractice: data.bestPractice,
                        status: data.status,
                        exclusionJustification:
                            data.status === "EXCLUDED"
                                ? data.exclusionJustification
                                : null,
                    },
                },
            });
        } else {
            // Create a new control
            await mutate({
                variables: {
                    input: {
                        frameworkId: props.frameworkId,
                        name: data.name,
                        description: data.description || null,
                        sectionTitle: data.sectionTitle,
                        bestPractice: data.bestPractice ?? true,
                        status: data.status,
                        exclusionJustification:
                            data.status === "EXCLUDED"
                                ? data.exclusionJustification
                                : null,
                    },
                    connections: [props.connectionId!],
                },
            });
            reset();
        }
        dialogRef.current?.close();
    });

    return (
        <Dialog
            trigger={props.children}
            ref={dialogRef}
            title={
                <Breadcrumb
                    items={[
                        __("Controls"),
                        frameworkControl
                            ? __("Edit Control")
                            : __("New Control"),
                    ]}
                />
            }
        >
            <form onSubmit={onSubmit}>
                <DialogContent padded className="space-y-2">
                    <Input
                        id="sectionTitle"
                        required
                        variant="ghost"
                        placeholder={__("Section title")}
                        {...register("sectionTitle")}
                    />
                    <Input
                        id="title"
                        required
                        variant="title"
                        placeholder={__("Document title")}
                        {...register("name")}
                    />
                    <Textarea
                        id="content"
                        variant="ghost"
                        autogrow
                        placeholder={__("Add description")}
                        {...register("description")}
                    />
                    <label className="flex items-center gap-2 cursor-pointer">
                        <Checkbox
                            checked={bestPracticeValue}
                            onChange={(checked) =>
                                setValue("bestPractice", checked)
                            }
                        />
                        <span className="text-sm">{__("Best Practice")}</span>
                    </label>
                </DialogContent>
                <DialogFooter>
                    <Button type="submit" disabled={isMutating}>
                        {props.control
                            ? __("Update control")
                            : __("Create control")}
                    </Button>
                </DialogFooter>
            </form>
        </Dialog>
    );
}
