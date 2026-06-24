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
        reportFile {
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
    reportFile {
      id
      alias
      isUserAuthorized
      access {
        id
        status
      }
    }
    framework {
      id
      name
      lightLogo {
        downloadUrl
      }
      darkLogo {
        downloadUrl
      }
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
  const reportPath = audit.reportFile?.alias ?? audit.reportFile?.id;
  const hasRequested = audit.reportFile?.access?.status === "REQUESTED";

  const [requestAccess, isRequestingAccess]
    = useMutation<AuditRow_requestAccessMutation>(requestAccessMutation);

  const handleRequestAccess = () => {
    requestAccess({
      variables: {
        input: {
          reportId: audit.reportFile?.id ?? "",
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
          searchParams.set("request-report-id", audit.reportFile?.id ?? "");
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
      {audit.reportFile && (
        audit.reportFile.isUserAuthorized
          ? (
              <Button
                className="w-full md:w-max"
                variant="secondary"
                icon={IconArrowLink}
                to={reportPath ? `/documents/${reportPath}` : undefined}
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
              lightLogoURL={audit.framework.lightLogo?.downloadUrl}
              darkLogoURL={audit.framework.darkLogo?.downloadUrl}
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
          lightLogoURL={audit.framework.lightLogo?.downloadUrl}
          darkLogoURL={audit.framework.darkLogo?.downloadUrl}
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
