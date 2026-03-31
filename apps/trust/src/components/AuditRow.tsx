// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { UnAuthenticatedError } from "@probo/relay";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  FrameworkLogo,
  IconArrowLink,
  IconLock,
  IconMedal,
  Table,
  useToast,
} from "@probo/ui";
import { type PropsWithChildren } from "react";
import { useFragment, useMutation } from "react-relay";
import { useLocation, useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import { getPathPrefix } from "#/utils/pathPrefix";

import type { AuditRow_requestAccessMutation } from "./__generated__/AuditRow_requestAccessMutation.graphql";
import type { AuditRowFragment$key } from "./__generated__/AuditRowFragment.graphql";

const requestAccessMutation = graphql`
  mutation AuditRow_requestAccessMutation($input: RequestReportAccessInput!) {
    requestReportAccess(input: $input) {
      audit {
        report {
          access {
            id
            status
          }
        }
      }
    }
  }
`;

const auditRowFragment = graphql`
  fragment AuditRowFragment on Audit {
    name
    report {
      id
      isUserAuthorized
      access {
        id
        status
      }
    }
    framework {
      id
      name
      lightLogoURL
      darkLogoURL
    }
  }
`;

export function AuditRow(props: { audit: AuditRowFragment$key }) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [searchParams] = useSearchParams();
  const location = useLocation();
  const navigate = useNavigate();

  const audit = useFragment(auditRowFragment, props.audit);
  const hasRequested = audit.report?.access?.status === "REQUESTED";

  const [requestAccess, isRequestingAccess]
    = useMutation<AuditRow_requestAccessMutation>(requestAccessMutation);

  const handleRequestAccess = () => {
    requestAccess({
      variables: {
        input: {
          reportId: audit.report?.id ?? "",
        },
      },
      onCompleted: (_, errors) => {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(__("Cannot request access"), errors),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Access request submitted successfully."),
          variant: "success",
        });
      },
      onError: (error) => {
        if (error instanceof UnAuthenticatedError) {
          const pathPrefix = getPathPrefix();
          searchParams.set("request-report-id", audit.report?.id ?? "");
          const urlSearchParams = new URLSearchParams([[
            "continue",
            window.location.origin + pathPrefix + location.pathname + "?" + searchParams.toString(),
          ]]);
          void navigate(`/connect?${urlSearchParams.toString()}`);

          return;
        }

        toast({
          title: __("Error"),
          description: error.message ?? __("Cannot request access"),
          variant: "error",
        });
      },
    });
  };

  return (
    <div className="text-sm border border-border-solid -mt-px flex gap-3 flex-col md:flex-row md:justify-between px-6 py-3">
      <div className="flex items-center gap-2">
        <IconMedal size={16} className="flex-none text-txt-tertiary" />
        {audit.name ?? audit.framework.name}
      </div>
      {audit.report && (
        audit.report.isUserAuthorized
          ? (
              <Button
                className="w-full md:w-max"
                variant="secondary"
                icon={IconArrowLink}
                to={`/documents/${audit.report.id}`}
              >
                {__("View")}
              </Button>
            )
          : (
              <Button
                disabled={hasRequested || isRequestingAccess}
                className="w-full md:w-max"
                variant="secondary"
                icon={IconLock}
                onClick={handleRequestAccess}
              >
                {hasRequested ? __("Access requested") : __("Request access")}
              </Button>
            )
      )}
    </div>
  );
}

export function AuditRowAvatar(props: { audit: AuditRowFragment$key }) {
  const audit = useFragment(auditRowFragment, props.audit);

  return (
    <>
      <AuditDialog audit={props.audit}>
        <button
          className="block cursor-pointer aspect-square"
          title={`Logo ${audit.framework.name}`}
        >
          <div className="flex flex-col gap-2 items-center w-19">
            <FrameworkLogo
              className="size-19"
              lightLogoURL={audit.framework.lightLogoURL}
              darkLogoURL={audit.framework.darkLogoURL}
              name={audit.framework.name}
            />
            <div className="txt-primary text-sm max-w-19 overflow-hidden min-w-0 whitespace-nowrap text-ellipsis">
              {audit.framework.name}
            </div>
          </div>
        </button>
      </AuditDialog>
    </>
  );
}

function AuditDialog(
  props: PropsWithChildren<{ audit: AuditRowFragment$key; logo?: string }>,
) {
  const audit = useFragment(auditRowFragment, props.audit);
  const location = useLocation();
  const { __ } = useTranslate();
  const items = [
    {
      label: __("Certifications"),
      to: location.pathname,
    },
    {
      label: audit.framework.name,
      to: location.pathname,
    },
  ];
  return (
    <Dialog
      trigger={props.children}
      className="max-w-[500px]"
      title={<Breadcrumb items={items} />}
    >
      <DialogContent className="p-4 lg:p-8 space-y-6">
        <FrameworkLogo
          className="size-24 mx-auto"
          lightLogoURL={audit.framework.lightLogoURL}
          darkLogoURL={audit.framework.darkLogoURL}
          name={audit.framework.name}
        />
        <h2 className="text-xl font-semibold mb-1">{audit.framework.name}</h2>
        <Table>
          <AuditRow audit={props.audit} />
        </Table>
      </DialogContent>
    </Dialog>
  );
}
