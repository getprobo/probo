import { formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { useToast } from "@probo/ui";
import { useEffect } from "react";
import { useMutation } from "react-relay";
import { useLocation, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import { getPathPrefix } from "#/utils/pathPrefix";

import type { useRequestAccessCallback_allMutation } from "./__generated__/useRequestAccessCallback_allMutation.graphql";
import type { useRequestAccessCallback_documentMutation } from "./__generated__/useRequestAccessCallback_documentMutation.graphql";
import type { useRequestAccessCallback_fileMutation } from "./__generated__/useRequestAccessCallback_fileMutation.graphql";
import type { useRequestAccessCallback_reportMutation } from "./__generated__/useRequestAccessCallback_reportMutation.graphql";

const documentMutation = graphql`
  mutation useRequestAccessCallback_documentMutation(
    $input: RequestDocumentAccessInput!
  ) {
    requestDocumentAccess(input: $input) {
      document {
        access {
          id
          status
        }
      }
    }
  }
`;

const reportMutation = graphql`
  mutation useRequestAccessCallback_reportMutation(
    $input: RequestReportAccessInput!
  ) {
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

const fileMutation = graphql`
  mutation useRequestAccessCallback_fileMutation(
    $input: RequestTrustCenterFileAccessInput!
  ) {
    requestTrustCenterFileAccess(input: $input) {
      file {
        access {
          id
          status
        }
      }
    }
  }
`;

const allMutation = graphql`
  mutation useRequestAccessCallback_allMutation {
    requestAllAccesses {
      trustCenterAccess {
        id
      }
    }
  }
`;

function errorToastArgs(__: (s: string) => string, error: GraphQLError | GraphQLError[]) {
  return {
    title: __("Error"),
    description: formatError(__("Cannot request access"), error),
    variant: "error" as const,
  };
}

function successToastArgs(__: (s: string) => string) {
  return {
    title: __("Success"),
    description: __("Access request submitted successfully."),
    variant: "success" as const,
  };
}

export function useRequestAccessCallback() {
  const [searchParams, setSearchParams] = useSearchParams();
  const location = useLocation();

  const documentId = searchParams.get("request-document-id");
  const reportId = searchParams.get("request-report-id");
  const fileId = searchParams.get("request-file-id");
  const all = searchParams.get("request-all");

  const { __ } = useTranslate();
  const { toast } = useToast();

  const [requestDocumentAccess] = useMutation<useRequestAccessCallback_documentMutation>(documentMutation);
  const [requestReportAccess] = useMutation<useRequestAccessCallback_reportMutation>(reportMutation);
  const [requestFileAccess] = useMutation<useRequestAccessCallback_fileMutation>(fileMutation);
  const [requestAll] = useMutation<useRequestAccessCallback_allMutation>(allMutation);

  useEffect(() => {
    if (documentId) {
      searchParams.delete("request-document-id");
      void requestDocumentAccess({
        variables: {
          input: { documentId },
        },
        onCompleted: (_, errors) => {
          if (errors?.length) {
            toast(errorToastArgs(__, errors));
            return;
          }

          toast(successToastArgs(__));
        },
        onError: (error) => {
          toast(errorToastArgs(__, error));
        },
      });
      setSearchParams(searchParams);
    } else if (reportId) {
      searchParams.delete("request-report-id");
      void requestReportAccess({
        variables: {
          input: { reportId },
        },
        onCompleted: (_, errors) => {
          if (errors?.length) {
            toast(errorToastArgs(__, errors));
            return;
          }

          toast(successToastArgs(__));
        },
        onError: (error) => {
          toast(errorToastArgs(__, error));
        },
      });
      setSearchParams(searchParams);
    } else if (fileId) {
      searchParams.delete("request-file-id");
      void requestFileAccess({
        variables: {
          input: { trustCenterFileId: fileId },
        },
        onCompleted: (_, errors) => {
          if (errors?.length) {
            toast(errorToastArgs(__, errors));
            return;
          }

          toast(successToastArgs(__));
        },
        onError: (error) => {
          toast(errorToastArgs(__, error));
        },
      });
      setSearchParams(searchParams);
    } else if (all) {
      searchParams.delete("request-all");
      void requestAll({
        variables: {},
        onCompleted: (_, errors) => {
          if (errors?.length) {
            toast(errorToastArgs(__, errors));
            return;
          }

          toast(successToastArgs(__));
          window.location.href = window.location.origin + getPathPrefix() + location.pathname;
        },
        onError: (error) => {
          toast(errorToastArgs(__, error));
        },
      });
      setSearchParams(searchParams);
    }
  }, [
    documentId,
    reportId,
    fileId,
    all,
    __,
    requestDocumentAccess,
    requestReportAccess,
    requestFileAccess,
    requestAll,
    searchParams,
    setSearchParams,
    toast,
    location,
  ]);
}
