# Forms and validation

Forms in the v2 apps are built on **Base UI** `Form` + `Field`, which extend the native [Constraint Validation API](https://developer.mozilla.org/en-US/docs/Web/HTML/Constraint_validation). The guiding rule matches the rest of the UI kit: **reach for the lightest tier that does the job**, and don't pile a form abstraction on top of the headless library when native validation already covers the case.

There is **no custom form hook.** The legacy `useFormWithSchema` (react-hook-form + zod wrapper) is not used in v2 — it forces every form, however trivial, onto react-hook-form + zod and hides which validation tier you are actually in.

## Related guides

| Topic | Guide |
|-------|--------|
| Component shape, `*Form` / `*Field` suffixes | [`contrib/claude/react-components.md`](react-components.md) |
| Styling primitives, Base UI | [`contrib/claude/ui.md`](ui.md) |
| Mutations, server errors, toasts | [`contrib/claude/relay.md`](relay.md) |
| Error/fallback handling | [`contrib/claude/error-handling.md`](error-handling.md) |
| Translating labels and messages | [`contrib/claude/i18n.md`](i18n.md) |

## The four tiers

Pick the **lowest** tier that expresses your validation. Do **not** mix tiers within one form.

| Tier | Tool | Use when |
|------|------|----------|
| 1 | Base UI `Field` + **native constraints** | The default. Validation is expressible as `required`, `type`, `minLength`, `maxLength`, `min`, `max`, `pattern`. |
| 2 | Base UI `Field` + **`validate` function** | A rule native constraints can't express: cross-field (confirm password), conditional, async uniqueness checks. |
| 3 | **zod** schema parsed in `onSubmit`, mapped to `Form` `errors` | Validation is genuinely complex, needs specific custom messages, or you want one typed schema as the source of truth (possibly shared). **No react-hook-form.** |
| 4 | **react-hook-form** + `zodResolver` | Large/dynamic forms: many fields, field arrays, wizards, heavy conditional logic, or perf with many controlled inputs. |

Server-side errors map onto the `Form` `errors` prop in **every** tier (see [Server errors](#server-errors)).

## Tier 1 — native constraints (default)

Most forms need nothing more. Base UI validates native HTML constraints, and `<Field.Error>` renders the message. Use `match` to supply your own copy (and i18n) per validity state.

```tsx
import { Form } from "@base-ui/react/form";
import { Field } from "@base-ui/react/field";
import { useTranslation } from "react-i18next";

export function CreateMeasureForm({ onSubmit }: CreateMeasureFormProps) {
  const { t } = useTranslation();
  return (
    <Form onSubmit={onSubmit}>
      <Field.Root name="name">
        <Field.Label>{t("measures.form.name")}</Field.Label>
        <Field.Control required maxLength={120} />
        <Field.Error match="valueMissing">{t("measures.form.nameRequired")}</Field.Error>
        <Field.Error match="tooLong">{t("measures.form.nameTooLong")}</Field.Error>
      </Field.Root>
      <Button type="submit">{t("common.save")}</Button>
    </Form>
  );
}
```

> `Field.Control` renders a native input by default; pass a Base UI input/select/checkbox via `render` (or nest one) when you need a styled control. The kit's styled inputs live under `form/` (see [`ui.md`](ui.md)).

## Tier 2 — custom `validate`

When the rule isn't a native constraint, add a sync/async `validate` on `Field.Root`. It returns a message string (or array) when invalid, or `null` when valid, and runs after native constraints pass. Stay in Base UI — do not introduce a library for this.

```tsx
<Field.Root
  name="confirmPassword"
  validate={(value, formValues) =>
    value === formValues.password ? null : t("auth.passwordsDoNotMatch")
  }
>
  <Field.Label>{t("auth.confirmPassword")}</Field.Label>
  <Field.Control type="password" required />
  <Field.Error />
</Field.Root>
```

Async `validate` (e.g. a uniqueness check) is supported; note Base UI does not block submit on a pending async validation in `validationMode="onSubmit"`.

## Tier 3 — zod, without react-hook-form

When validation is complex or needs specific messages, parse a zod schema in `onFormSubmit` and feed the flattened field errors to `Form`'s `errors` prop. This keeps a single typed schema as the source of truth **without** pulling in react-hook-form.

```tsx
import { Form } from "@base-ui/react/form";
import { z } from "zod";

const schema = z.object({
  name: z.string().min(1, "measures.form.nameRequired"),
  weight: z.coerce.number().min(0).max(100),
});

export function CreateMeasureForm({ onValid }: CreateMeasureFormProps) {
  const [errors, setErrors] = useState<Record<string, string | string[]>>({});

  return (
    <Form
      errors={errors}
      onClearErrors={setErrors}
      onFormSubmit={(formData) => {
        const result = schema.safeParse(Object.fromEntries(formData));
        if (!result.success) {
          setErrors(z.flattenError(result.error).fieldErrors);
          return;
        }
        onValid(result.data);
      }}
    >
      <Field.Root name="name">
        <Field.Label>{t("measures.form.name")}</Field.Label>
        <Field.Control />
        <Field.Error />
      </Field.Root>
      {/* … */}
    </Form>
  );
}
```

Reserve zod for the cases above — a `z.object({ name: z.string().min(1) })` that only restates a `required` attribute belongs in tier 1.

## Tier 4 — react-hook-form

Only when the form is large or dynamic enough that react-hook-form's machinery pays for itself. Use `useForm` + `zodResolver` **directly** (no project wrapper hook), and wire Base UI `Field.Control` through `register` / `Controller`.

```tsx
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";

const { register, handleSubmit, formState: { errors } } = useForm({
  resolver: zodResolver(schema),
});
```

If you find yourself reaching for tier 4 for an ordinary CRUD dialog, step back — it is almost always tier 1 or 3.

## Server errors

Validation that only the backend can perform (uniqueness, business rules — see [`validation.md`](validation.md) for the error codes the API returns) comes back from the mutation. Map it onto the `Form` `errors` prop keyed by field `name`; Base UI clears each entry when the field changes. Use `formatError` from `@probo/helpers` to turn a GraphQL error into a message.

```tsx
const toast = Toast.useToastManager();
const [createMeasure, isCreating] = useMutation<CreateMeasureMutation>(createMeasureMutation);
const [errors, setErrors] = useState<Record<string, string | string[]>>({});

function onValid(input: MeasureInput) {
  createMeasure({
    variables: { input, connections: [connectionId] },
    onCompleted() {
      toast.add({ title: t("measures.created"), type: "success" });
    },
    onError(error) {
      // Field-specific server errors → the form; otherwise → a toast.
      const fieldErrors = toFieldErrors(error);
      if (fieldErrors) {
        setErrors(fieldErrors);
      } else {
        toast.add({
          title: t("common.error"),
          description: formatError(t("measures.createFailed"), error as GraphQLError),
          type: "error",
        });
      }
    },
  });
}
```

- **Field-level** server errors → `Form` `errors`.
- **Whole-form / unexpected** errors → a toast (see [`error-handling.md`](error-handling.md) and [`ui.md`](ui.md)).
- Never rely on a page reload to surface a failed submit.

## Submit, loading, and disabled state

- Disable the submit control while the mutation is in flight (`isCreating` / `isUpdating` — see the mutation naming in [`relay.md`](relay.md)).
- Keep the form responsive: do not unmount fields on submit; let Base UI manage validity.
- For destructive submits, wrap in `useConfirm` (see [`relay.md`](relay.md)).

## Dialogs

Forms in dialogs use the Base UI dialog directly — controlled with `open` / `onOpenChange`. **Do not** use the legacy `useDialogRef` / imperative `open()` pattern (see [`ui.md`](ui.md#headless-primitives-style-dont-re-implement)).

```tsx
// Bad — imperative ref to open a form dialog (legacy console pattern)
const ref = useDialogRef();
ref.current?.open();

// Good — controlled open state, same API as the Base UI dialog
const [open, setOpen] = useState(false);
<Dialog open={open} onOpenChange={setOpen}>{/* … <CreateMeasureForm /> … */}</Dialog>
```

## Do / don't summary

```text
// Bad
- Custom useFormWithSchema wrapper around every form
- zod schema that only restates `required` / `maxLength`
- react-hook-form for a two-field dialog
- Mixing native constraints and a zod parse in the same form
- Opening a form dialog through an imperative ref

// Good
- Native Base UI constraints by default
- `validate` for cross-field / async rules
- zod (no RHF) when validation is genuinely complex
- react-hook-form only for large / dynamic forms
- Server errors mapped onto the Form `errors` prop
```
