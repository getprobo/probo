// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { Form } from "@base-ui/react/form";
import { XIcon } from "@phosphor-icons/react";
import { toFieldErrors } from "@probo/helpers";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Checkbox } from "@probo/ui/src/v2/Checkbox/Checkbox";
import { Drawer } from "@probo/ui/src/v2/Drawer/Drawer";
import { DrawerBody } from "@probo/ui/src/v2/Drawer/DrawerBody";
import { DrawerClose } from "@probo/ui/src/v2/Drawer/DrawerClose";
import { DrawerDescription } from "@probo/ui/src/v2/Drawer/DrawerDescription";
import { DrawerFooter } from "@probo/ui/src/v2/Drawer/DrawerFooter";
import { DrawerHeader } from "@probo/ui/src/v2/Drawer/DrawerHeader";
import { DrawerPopup } from "@probo/ui/src/v2/Drawer/DrawerPopup";
import { DrawerTitle } from "@probo/ui/src/v2/Drawer/DrawerTitle";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Textarea } from "@probo/ui/src/v2/form/Textarea";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type {
  CategoryDrawer_cookieCategory$data,
  CategoryDrawer_cookieCategory$key,
} from "#/__generated__/core/CategoryDrawer_cookieCategory.graphql";
import type { CategoryDrawerUpdateMutation } from "#/__generated__/core/CategoryDrawerUpdateMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { categoryDrawer } from "../../../variants";
import { GCM_CONSENT_TYPES, setIncluded, TCF_PURPOSE_IDS } from "../_lib/categoryConsent";
import {
  CATEGORY_DESCRIPTION_MAX_LENGTH,
  CATEGORY_NAME_MAX_LENGTH,
  CATEGORY_SLUG_MAX_LENGTH,
  CATEGORY_SLUG_PATTERN,
} from "../_lib/categorySlug";

const fragment = graphql`
  fragment CategoryDrawer_cookieCategory on CookieCategory {
    id
    name
    slug
    description
    kind
    gcmConsentTypes
    tcfPurposeIds
    posthogConsent
  }
`;

const updateMutation = graphql`
  mutation CategoryDrawerUpdateMutation($input: UpdateCookieCategoryInput!) {
    updateCookieCategory(input: $input) {
      cookieCategory {
        id
        name
        slug
        description
        gcmConsentTypes
        tcfPurposeIds
        posthogConsent
        updatedAt
      }
      cookieBanner {
        id
        categories(first: 100, orderBy: { field: RANK, direction: ASC }, filter: { excludeKind: UNCATEGORISED }) {
          edges {
            node {
              id
              posthogConsent
            }
          }
        }
        latestVersion {
          id
          version
          state
        }
      }
    }
  }
`;

interface CategoryDrawerProps {
  categoryKey: CategoryDrawer_cookieCategory$key;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CategoryDrawer({
  categoryKey,
  open,
  onOpenChange,
}: CategoryDrawerProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const category = useFragment(fragment, categoryKey);

  return (
    <Drawer open={open} onOpenChange={onOpenChange} swipeDirection="right">
      <DrawerPopup side="right" size={2}>
        <DrawerHeader>
          <DrawerTitle>{t("categoryDrawer.title")}</DrawerTitle>
          <DrawerClose
            render={(
              <IconButton
                size={1}
                variant="ghost"
                color="neutral"
                aria-label={t("categoryDrawer.actions.close")}
              >
                <XIcon />
              </IconButton>
            )}
          />
        </DrawerHeader>
        <DrawerDescription>
          {t("categoryDrawer.description")}
        </DrawerDescription>
        {open && (
          <CategoryDrawerForm
            category={category}
            onClose={() => onOpenChange(false)}
          />
        )}
      </DrawerPopup>
    </Drawer>
  );
}

interface CategoryDrawerFormProps {
  category: CategoryDrawer_cookieCategory$data;
  onClose: () => void;
}

function CategoryDrawerForm({ category, onClose }: CategoryDrawerFormProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { form, fields, field, checks, check, checkLabel, footerActions } = categoryDrawer();
  const [name, setName] = useState(category.name);
  const [slug, setSlug] = useState(category.slug);
  const [description, setDescription] = useState(category.description);
  const [gcmConsentTypes, setGcmConsentTypes] = useState<string[]>(
    [...category.gcmConsentTypes],
  );
  const [tcfPurposeIds, setTcfPurposeIds] = useState<number[]>(
    [...category.tcfPurposeIds],
  );
  const [posthogConsent, setPosthogConsent] = useState(category.posthogConsent);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [updateCategory, isUpdating] = useMutation<CategoryDrawerUpdateMutation>(
    updateMutation,
    {
      successMessage: t("categoryDrawer.messages.updated"),
      errorToast: t("categoryDrawer.errors.update"),
    },
  );

  function handleSubmit() {
    void updateCategory({
      variables: {
        input: {
          cookieCategoryId: category.id,
          name: name.trim(),
          slug: slug.trim(),
          description: description.trim(),
          gcmConsentTypes,
          tcfPurposeIds: [...tcfPurposeIds].sort((left, right) => left - right),
          posthogConsent: category.kind === "NORMAL" ? posthogConsent : false,
        },
      },
      onCompleted(_response, payloadErrors) {
        const fieldErrors = toFieldErrors(payloadErrors);
        if (fieldErrors != null) {
          setErrors(fieldErrors);
        }
      },
    }).then(
      () => {
        onClose();
      },
      () => {
        // Field errors are mapped in onCompleted; other failures toast.
      },
    );
  }

  return (
    <Form className={form()} errors={errors} onFormSubmit={handleSubmit}>
      <DrawerBody>
        <div className={fields()}>
          <Field required label={t("categoryDrawer.fields.name")} error={errors.name}>
            <TextField
              name="name"
              required
              maxLength={CATEGORY_NAME_MAX_LENGTH}
              value={name}
              disabled={isUpdating}
              onValueChange={(value) => {
                setName(value);
                setErrors({});
              }}
            />
          </Field>
          <div className={field()}>
            <Field required label={t("categoryDrawer.fields.slug")} error={errors.slug}>
              <TextField
                name="slug"
                required
                maxLength={CATEGORY_SLUG_MAX_LENGTH}
                pattern={CATEGORY_SLUG_PATTERN}
                value={slug}
                disabled={isUpdating}
                onValueChange={(value) => {
                  setSlug(value);
                  setErrors({});
                }}
              />
            </Field>
            <Text size={1} color="faint">
              {t("categoryDrawer.fields.slugHelp")}
            </Text>
          </div>
          <Field
            required
            label={t("categoryDrawer.fields.description")}
            error={errors.description}
          >
            <Textarea
              name="description"
              required
              maxLength={CATEGORY_DESCRIPTION_MAX_LENGTH}
              rows={3}
              value={description}
              disabled={isUpdating}
              onChange={(event) => {
                setDescription(event.target.value);
                setErrors({});
              }}
            />
          </Field>
          <div className={checks()}>
            <Text size={2} weight="medium" highContrast>
              {t("categoryDrawer.googleConsentMode.title")}
            </Text>
            <Text size={1} color="faint">
              {t("categoryDrawer.googleConsentMode.description")}
            </Text>
            {GCM_CONSENT_TYPES.map(type => (
              <label key={type} className={check()}>
                <Checkbox
                  checked={gcmConsentTypes.includes(type)}
                  disabled={isUpdating}
                  onCheckedChange={(checked) => {
                    setGcmConsentTypes(setIncluded(gcmConsentTypes, type, checked === true));
                  }}
                />
                <Text size={1} className={checkLabel()}>
                  {type}
                </Text>
              </label>
            ))}
          </div>
          <div className={checks()}>
            <Text size={2} weight="medium" highContrast>
              {t("categoryDrawer.tcfPurposes.title")}
            </Text>
            <Text size={1} color="faint">
              {t("categoryDrawer.tcfPurposes.description")}
            </Text>
            {TCF_PURPOSE_IDS.map(id => (
              <label key={id} className={check()}>
                <Checkbox
                  checked={tcfPurposeIds.includes(id)}
                  disabled={isUpdating}
                  onCheckedChange={(checked) => {
                    setTcfPurposeIds(
                      setIncluded(tcfPurposeIds, id, checked === true)
                        .sort((left, right) => left - right),
                    );
                  }}
                />
                <Text size={1} className={checkLabel()}>
                  {id}
                </Text>
              </label>
            ))}
          </div>
          {category.kind === "NORMAL" && (
            <div className={checks()}>
              <Text size={2} weight="medium" highContrast>
                {t("categoryDrawer.posthog.title")}
              </Text>
              <Text size={1} color="faint">
                {t("categoryDrawer.posthog.description")}
              </Text>
              <label className={check()}>
                <Checkbox
                  checked={posthogConsent}
                  disabled={isUpdating}
                  onCheckedChange={(checked) => {
                    setPosthogConsent(checked === true);
                  }}
                />
                <Text size={2}>
                  {t("categoryDrawer.posthog.checkbox")}
                </Text>
              </label>
            </div>
          )}
        </div>
      </DrawerBody>
      <DrawerFooter>
        <div className={footerActions()}>
          <Button
            type="button"
            variant="soft"
            color="neutral"
            onClick={onClose}
          >
            {t("categoryDrawer.actions.cancel")}
          </Button>
          <Button type="submit" variant="solid" color="neutral" highContrast loading={isUpdating}>
            {t("categoryDrawer.actions.save")}
          </Button>
        </div>
      </DrawerFooter>
    </Form>
  );
}
