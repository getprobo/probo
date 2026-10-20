// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import {
  Badge,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  IconMagnifyingGlass,
  IconPlusLarge,
  IconTrashCan,
  InfiniteScrollTrigger,
  Input,
  Option,
  Select,
  Spinner,
} from "@probo/ui";
import { type ReactNode, Suspense, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";

import { usePaginatedInternalControls } from "#/hooks/graph/usePaginatedInternalControls";
import { useOrganizationId } from "#/hooks/useOrganizationId";

type Props = {
  children: ReactNode;
  connectionId: string;
  disabled?: boolean;
  linkedInternalControls?: { id: string }[];
  onLink: (internalControlId: string) => void;
  onUnlink: (internalControlId: string) => void;
};

export function LinkedInternalControlDialog({ children, ...props }: Props) {
  const { t } = useTranslation();

  return (
    <Dialog trigger={children} title={t("linkedInternalControlsDialog.title")}>
      <DialogContent>
        <Suspense fallback={<Spinner centered />}>
          <LinkedInternalControlsDialogContent {...props} />
        </Suspense>
      </DialogContent>
      <DialogFooter exitLabel={t("linkedInternalControlsDialog.actions.close")} />
    </Dialog>
  );
}

function LinkedInternalControlsDialogContent(props: Omit<Props, "children">) {
  const organizationId = useOrganizationId();
  const { data, loadNext, hasNext, isLoadingNext }
    = usePaginatedInternalControls(organizationId);
  const { t } = useTranslation();
  const [search, setSearch] = useState("");
  const [category, setCategory] = useState<string | null>(null);
  const internalControls = useMemo(
    () => data.internalControls?.edges?.map(edge => edge.node) ?? [],
    [data.internalControls],
  );
  const linkedIds = useMemo(() => {
    return new Set(props.linkedInternalControls?.map(m => m.id) ?? []);
  }, [props.linkedInternalControls]);

  const filteredInternalControls = useMemo(() => {
    return internalControls.filter(
      internalControl =>
        (category === null || internalControl.category === category)
        && (internalControl.name.toLowerCase().includes(search.toLowerCase())
          || internalControl.description?.toLowerCase().includes(search.toLowerCase())),
    );
  }, [internalControls, search, category]);

  const categories = useMemo(
    () => Array.from(new Set(internalControls.map(m => m.category))),
    [internalControls],
  );

  return (
    <>
      <div className="flex items-center gap-2 sticky top-0 relative py-4 bg-linear-to-b from-50% from-level-2 to-level-2/0 px-6">
        <Input
          icon={IconMagnifyingGlass}
          placeholder={t("linkedInternalControlsDialog.searchPlaceholder")}
          onValueChange={setSearch}
        />
        <Select
          value={category ?? ""}
          placeholder={t("linkedInternalControlsDialog.allCategories")}
          onValueChange={setCategory}
          className="max-w-[180px]"
        >
          {categories.map(category => (
            <Option key={category} value={category}>
              {category}
            </Option>
          ))}
        </Select>
      </div>
      <div className="divide-y divide-border-low">
        {filteredInternalControls.map(internalControl => (
          <InternalControlRow
            key={internalControl.id}
            internalControl={internalControl}
            linkedInternalControls={linkedIds}
            onLink={props.onLink}
            onUnlink={props.onUnlink}
            disabled={props.disabled}
          />
        ))}
        {hasNext && (
          <InfiniteScrollTrigger
            loading={isLoadingNext}
            onView={() => loadNext(20)}
          />
        )}
      </div>
    </>
  );
}

type RowProps = {
  internalControl: { name: string; category: string; id: string };
  linkedInternalControls: Set<string>;
  disabled?: boolean;
  onLink: (internalControlId: string) => void;
  onUnlink: (internalControlId: string) => void;
};

function InternalControlRow(props: RowProps) {
  const { t } = useTranslation();

  const isLinked = props.linkedInternalControls.has(props.internalControl.id);
  const onClick = isLinked ? props.onUnlink : props.onLink;
  const IconComponent = isLinked ? IconTrashCan : IconPlusLarge;

  return (
    <button
      className="py-4 flex items-center gap-4 hover:bg-subtle cursor-pointer px-6 w-full"
      onClick={() => onClick(props.internalControl.id)}
    >
      {props.internalControl.name}
      <Badge variant="neutral">{props.internalControl.category}</Badge>
      <Button
        disabled={props.disabled}
        className="ml-auto"
        variant={isLinked ? "secondary" : "primary"}
        asChild
      >
        <span>
          <IconComponent size={16} />
          {" "}
          {isLinked
            ? t("linkedInternalControlsDialog.actions.unlink")
            : t("linkedInternalControlsDialog.actions.link")}
        </span>
      </Button>
    </button>
  );
}
