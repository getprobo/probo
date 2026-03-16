import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { UnAuthenticatedError } from "@probo/relay";
import {
  Button,
  IconArrowLink,
  IconLock,
  IconPageTextLine,
  useToast,
} from "@probo/ui";
import { useFragment, useMutation } from "react-relay";
import { useLocation, useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import { getPathPrefix } from "#/utils/pathPrefix";

import type { DocumentRow_requestAccessMutation } from "./__generated__/DocumentRow_requestAccessMutation.graphql";
import type { DocumentRowFragment$key } from "./__generated__/DocumentRowFragment.graphql";

const requestAccessMutation = graphql`
  mutation DocumentRow_requestAccessMutation(
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

const documentRowFragment = graphql`
  fragment DocumentRowFragment on Document {
    id
    title
    isUserAuthorized
    access {
      id
      status
    }
  }
`;

export function DocumentRow(props: { document: DocumentRowFragment$key }) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();

  const document = useFragment(documentRowFragment, props.document);
  const hasRequested = document.access?.status === "REQUESTED";

  const [requestAccess, isRequestingAccess]
    = useMutation<DocumentRow_requestAccessMutation>(requestAccessMutation);

  const handleRequestAccess = () => {
    requestAccess({
      variables: {
        input: {
          documentId: document.id,
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
          searchParams.set("request-document-id", document.id);
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
        <IconPageTextLine size={16} className=" flex-none text-txt-tertiary" />
        {document.title}
      </div>
      {document.isUserAuthorized
        ? (
            <Button
              className="w-full md:w-max"
              variant="secondary"
              icon={IconArrowLink}
              onClick={() => void navigate(`/documents/${document.id}`)}
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
          )}
    </div>
  );
}
