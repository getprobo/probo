/**
 * @generated SignedSource<<7c575bd72e0d697cf92b9dd57038995c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type TrustCenterAccessState = "ACTIVE" | "INACTIVE";
export type TrustCenterDocumentAccessStatus = "GRANTED" | "REJECTED" | "REQUESTED" | "REVOKED";
export type UpdateTrustCenterAccessInput = {
  documents?: ReadonlyArray<TrustCenterDocumentAccessInput> | null | undefined;
  id: string;
  name?: string | null | undefined;
  reports?: ReadonlyArray<TrustCenterDocumentAccessInput> | null | undefined;
  state?: TrustCenterAccessState | null | undefined;
  trustCenterFiles?: ReadonlyArray<TrustCenterDocumentAccessInput> | null | undefined;
};
export type TrustCenterDocumentAccessInput = {
  id: string;
  status: TrustCenterDocumentAccessStatus;
};
export type CompliancePageAccessEditDialogUpdateMutation$variables = {
  input: UpdateTrustCenterAccessInput;
};
export type CompliancePageAccessEditDialogUpdateMutation$data = {
  readonly updateTrustCenterAccess: {
    readonly trustCenterAccess: {
      readonly activeCount: number;
      readonly createdAt: string;
      readonly email: string;
      readonly hasAcceptedNonDisclosureAgreement: boolean;
      readonly id: string;
      readonly name: string;
      readonly pendingRequestCount: number;
      readonly state: TrustCenterAccessState;
      readonly updatedAt: string;
    };
  };
};
export type CompliancePageAccessEditDialogUpdateMutation = {
  response: CompliancePageAccessEditDialogUpdateMutation$data;
  variables: CompliancePageAccessEditDialogUpdateMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "UpdateTrustCenterAccessPayload",
    "kind": "LinkedField",
    "name": "updateTrustCenterAccess",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "TrustCenterAccess",
        "kind": "LinkedField",
        "name": "trustCenterAccess",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "email",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "name",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "state",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "hasAcceptedNonDisclosureAgreement",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "createdAt",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "updatedAt",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "pendingRequestCount",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "activeCount",
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "CompliancePageAccessEditDialogUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "CompliancePageAccessEditDialogUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "e8401f17211bce414eee9048b939235a",
    "id": null,
    "metadata": {},
    "name": "CompliancePageAccessEditDialogUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageAccessEditDialogUpdateMutation(\n  $input: UpdateTrustCenterAccessInput!\n) {\n  updateTrustCenterAccess(input: $input) {\n    trustCenterAccess {\n      id\n      email\n      name\n      state\n      hasAcceptedNonDisclosureAgreement\n      createdAt\n      updatedAt\n      pendingRequestCount\n      activeCount\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "e511d9060cc29db26cd6ff281fe2722b";

export default node;
