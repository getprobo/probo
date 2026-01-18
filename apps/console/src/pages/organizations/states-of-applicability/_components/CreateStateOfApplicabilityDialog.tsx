import { useTranslate } from "@probo/i18n";
import {
    Breadcrumb,
    Button,
    Dialog,
    DialogContent,
    DialogFooter,
    Field,
    useDialogRef,
} from "@probo/ui";
import type { ReactNode } from "react";
import { useNavigate } from "react-router";
import { useOrganizationId } from "/hooks/useOrganizationId";
import z from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { graphql } from "react-relay";
import { PeopleSelectField } from "/components/form/PeopleSelectField";
import { Suspense } from "react";
import type { CreateStateOfApplicabilityDialogMutation } from "/__generated__/core/CreateStateOfApplicabilityDialogMutation.graphql";

const createStateOfApplicabilityMutation = graphql`
    mutation CreateStateOfApplicabilityDialogMutation(
        $input: CreateStateOfApplicabilityInput!
        $connections: [ID!]!
    ) {
        createStateOfApplicability(input: $input) {
            stateOfApplicabilityEdge @prependEdge(connections: $connections) {
                node {
                    id
                    name
                    createdAt
                    updatedAt
                }
            }
        }
    }
`;

type Props = {
    children: ReactNode;
    connectionId: string;
};

const schema = z.object({
    name: z.string().min(1),
    ownerId: z.string().min(1),
});

export function CreateStateOfApplicabilityDialog({
    children,
    connectionId,
}: Props) {
    const { __ } = useTranslate();
    const organizationId = useOrganizationId();
    const navigate = useNavigate();
    const { control, register, handleSubmit, reset } = useFormWithSchema(
        schema,
        {
            defaultValues: {
                name: "",
                ownerId: "",
            },
        },
    );
    const ref = useDialogRef();

    const [mutate, isMutating] =
        useMutationWithToasts<CreateStateOfApplicabilityDialogMutation>(
            createStateOfApplicabilityMutation,
            {
                successMessage: __(
                    "State of applicability created successfully.",
                ),
                errorMessage: __("Failed to create state of applicability"),
            },
        );

    const onSubmit = handleSubmit((data) => {
        mutate({
            variables: {
                input: {
                    name: data.name,
                    organizationId,
                    ownerId: data.ownerId,
                },
                connections: [connectionId],
            },
            onCompleted: (response) => {
                reset();
                ref.current?.close();
                const stateOfApplicabilityId =
                    response.createStateOfApplicability.stateOfApplicabilityEdge
                        .node.id;
                navigate(
                    `/organizations/${organizationId}/states-of-applicability/${stateOfApplicabilityId}`,
                );
            },
        });
    });

    return (
        <Dialog
            ref={ref}
            trigger={children}
            title={
                <Breadcrumb
                    items={[
                        __("States of Applicability"),
                        __("New State of Applicability"),
                    ]}
                />
            }
        >
            <form onSubmit={onSubmit}>
                <DialogContent padded className="space-y-4">
                    <Field
                        label={__("Name")}
                        {...register("name")}
                        type="text"
                        required
                    />
                    <Field label={__("Owner")}>
                        <Suspense fallback={<div>{__("Loading...")}</div>}>
                            <PeopleSelectField
                                organizationId={organizationId}
                                control={control}
                                name="ownerId"
                            />
                        </Suspense>
                    </Field>
                </DialogContent>
                <DialogFooter>
                    <Button disabled={isMutating} type="submit">
                        {__("Create")}
                    </Button>
                </DialogFooter>
            </form>
        </Dialog>
    );
}
