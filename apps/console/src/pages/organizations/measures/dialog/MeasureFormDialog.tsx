import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Input,
  Label,
  Option,
  PropertyRow,
  useDialogRef,
  type DialogRef,
} from "@probo/ui";
import type { ReactNode } from "react";
import { useTranslate } from "@probo/i18n";
import { Breadcrumb } from "@probo/ui";
import { graphql } from "relay-runtime";
import type { MeasureFormDialogMeasureFragment$key } from "./__generated__/MeasureFormDialogMeasureFragment.graphql";
import { useFragment } from "react-relay";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { getMeasureStateLabel, measureStates } from "@probo/helpers";
import { ControlledSelect } from "/components/form/ControlledField";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { useUpdateMeasure } from "/hooks/graph/MeasureGraph";

const measureFragment = graphql`
  fragment MeasureFormDialogMeasureFragment on Measure {
    id
    description
    name
    category
    state
  }
`;

const measureCreateMutation = graphql`
  mutation MeasureFormDialogCreateMutation(
    $input: CreateMeasureInput!
    $connections: [ID!]!
  ) {
    createMeasure(input: $input) {
      measureEdge @prependEdge(connections: $connections) {
        node {
          ...MeasureFormDialogMeasureFragment
        }
      }
    }
  }
`;

const measureSchema = z.object({
  name: z.string().min(1, "Name is required"),
  description: z.string().min(1, "Description is required"),
  category: z.string().min(1, "Category is required"),
  state: z.enum(measureStates),
});

type Props = {
  children?: ReactNode;
  measure?: MeasureFormDialogMeasureFragment$key;
  connection?: string;
  ref?: DialogRef;
};

export default function MeasureFormDialog(props: Props) {
  const { __ } = useTranslate();
  const measure = useFragment(measureFragment, props.measure);
  const dialogRef = props.ref ?? useDialogRef();
  const organizationId = useOrganizationId();
  const [mutate] = props.measure
    ? useUpdateMeasure()
    : useMutationWithToasts(measureCreateMutation, {
        successMessage: __("Measure created successfully."),
        errorMessage: __("Failed to create measure"),
      });

  const { control, handleSubmit, register, formState, reset } =
    useFormWithSchema(measureSchema, {
      values: {
        name: measure?.name ?? "",
        description: measure?.description ?? "",
        category: measure?.category ?? "",
        state: measure?.state ?? "NOT_STARTED",
      },
    });

  const onSubmit = handleSubmit(async (data) => {
    if (measure) {
      await mutate({
        variables: {
          input: {
            id: measure.id,
            name: data.name,
            description: data.description,
            category: data.category,
            state: data.state,
          },
        },
      });
    } else {
      await mutate({
        variables: {
          input: {
            organizationId,
            name: data.name,
            description: data.description,
            category: data.category,
          },
          connections: [props.connection!],
        },
      });
      reset();
    }
    dialogRef.current?.close();
  });

  return (
    <Dialog
      ref={dialogRef}
      trigger={props.children}
      title={
        <Breadcrumb
          items={[
            __("Measures"),
            measure ? __("Edit Measure") : __("New Measure"),
          ]}
        />
      }
    >
      <form onSubmit={onSubmit}>
        <DialogContent className="grid grid-cols-[1fr_420px]">
          <div className="py-8 px-10 space-y-6">
            <Field
              {...register("name")}
              error={formState.errors.name?.message}
              label={__("Measure name")}
              placeholder={__("Measure title")}
              required
            />
            <Field
              {...register("description")}
              error={formState.errors.description?.message}
              label={__("Description")}
              placeholder={__("Add description")}
              type="textarea"
              required
            />
          </div>
          {/* Properties form */}
          <div className="py-5 px-6 bg-subtle">
            <Label>{__("Properties")}</Label>
            <PropertyRow
              label={__("Category")}
              error={formState.errors.category?.message}
            >
              <Input
                {...register("category")}
                required
                placeholder={__("Select category")}
              />
            </PropertyRow>
            {measure && (
              <PropertyRow
                label={__("State")}
                error={formState.errors.state?.message}
              >
                <ControlledSelect
                  control={control}
                  name="state"
                  placeholder={__("Select state")}
                >
                  {measureStates.map((state) => (
                    <Option key={state} value={state}>
                      {getMeasureStateLabel(__, state)}
                    </Option>
                  ))}
                </ControlledSelect>
              </PropertyRow>
            )}
          </div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit">
            {measure ? __("Update measure") : __("Create measure")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
