/**
 * @generated SignedSource<<b9fc06fe9a0b5d316298bb77804e29b1>>
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
      readonly id: string;
      readonly pendingRequestCount: number;
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
    "cacheID": "39cf57753ab2bd46878a9200ef892894",
    "id": null,
    "metadata": {},
    "name": "CompliancePageAccessEditDialogUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageAccessEditDialogUpdateMutation(\n  $input: UpdateTrustCenterAccessInput!\n) {\n  updateTrustCenterAccess(input: $input) {\n    trustCenterAccess {\n      id\n      createdAt\n      updatedAt\n      pendingRequestCount\n      activeCount\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "896760c6c5c2d6adfca8ffcd7c7e22c7";

export default node;
