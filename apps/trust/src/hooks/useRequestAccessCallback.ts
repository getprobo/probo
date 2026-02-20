import { formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { useToast } from "@probo/ui";
import { useEffect } from "react";
import { useMutation } from "react-relay";
import { useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import type { useRequestAccessCallback_allMutation } from "./__generated__/useRequestAccessCallback_allMutation.graphql";
import type { useRequestAccessCallback_documentMutation } from "./__generated__/useRequestAccessCallback_documentMutation.graphql";
import type { useRequestAccessCallback_fileMutation } from "./__generated__/useRequestAccessCallback_fileMutation.graphql";
import type { useRequestAccessCallback_reportMutation } from "./__generated__/useRequestAccessCallback_reportMutation.graphql";

const documentMutation = graphql`
  mutation useRequestAccessCallback_documentMutation(
    $input: RequestDocumentAccessInput!
  ) {
    requestDocumentAccess(input: $input) {
      trustCenterAccess {
        id
      }
    }
  }
`;

const reportMutation = graphql`
  mutation useRequestAccessCallback_reportMutation(
    $input: RequestReportAccessInput!
  ) {
    requestReportAccess(input: $input) {
      trustCenterAccess {
        id
      }
    }
  }
`;

const fileMutation = graphql`
  mutation useRequestAccessCallback_fileMutation(
    $input: RequestTrustCenterFileAccessInput!
  ) {
    requestTrustCenterFileAccess(input: $input) {
      trustCenterAccess {
        id
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
      void requestDocumentAccess({
        variables: {
          input: { documentId },
        },
        onCompleted: (_, errors) => {
          if (errors?.length) {
            toast(errorToastArgs(__, errors));
            searchParams.delete("request-document-id");
            setSearchParams(searchParams);
            return;
          }

          toast(successToastArgs(__));
          searchParams.delete("request-document-id");
          setSearchParams(searchParams);
        },
        onError: (error) => {
          toast(errorToastArgs(__, error));
          searchParams.delete("request-document-id");
          setSearchParams(searchParams);
        },
      });
    } else if (reportId) {
      void requestReportAccess({
        variables: {
          input: { reportId },
        },
        onCompleted: (_, errors) => {
          if (errors?.length) {
            toast(errorToastArgs(__, errors));
            searchParams.delete("request-report-id");
            setSearchParams(searchParams);
            return;
          }

          toast(successToastArgs(__));
          searchParams.delete("request-report-id");
          setSearchParams(searchParams);
        },
        onError: (error) => {
          toast(errorToastArgs(__, error));
          searchParams.delete("request-report-id");
          setSearchParams(searchParams);
        },
      });
    } else if (fileId) {
      void requestFileAccess({
        variables: {
          input: { trustCenterFileId: fileId },
        },
        onCompleted: (_, errors) => {
          if (errors?.length) {
            toast(errorToastArgs(__, errors));
            searchParams.delete("request-file-id");
            setSearchParams(searchParams);
            return;
          }

          toast(successToastArgs(__));
          searchParams.delete("request-file-id");
          setSearchParams(searchParams);
        },
        onError: (error) => {
          toast(errorToastArgs(__, error));
          searchParams.delete("request-file-id");
          setSearchParams(searchParams);
        },
      });
    } else if (all) {
      void requestAll({
        variables: {},
        onCompleted: (_, errors) => {
          if (errors?.length) {
            toast(errorToastArgs(__, errors));
            searchParams.delete("request-all");
            setSearchParams(searchParams);
            return;
          }

          toast(successToastArgs(__));
          searchParams.delete("request-all");
          setSearchParams(searchParams);
        },
        onError: (error) => {
          toast(errorToastArgs(__, error));
          searchParams.delete("request-all");
          setSearchParams(searchParams);
        },
      });
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
  ]);
}
