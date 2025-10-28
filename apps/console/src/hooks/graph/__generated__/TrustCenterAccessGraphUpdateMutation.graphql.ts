/**
 * @generated SignedSource<<25b79c235754d47097b2ca28f1f3e9a0>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateTrustCenterAccessInput = {
  active?: boolean | null | undefined;
  documentIds?: ReadonlyArray<string> | null | undefined;
  id: string;
  name?: string | null | undefined;
  reportIds?: ReadonlyArray<string> | null | undefined;
  trustCenterFileIds?: ReadonlyArray<string> | null | undefined;
};
export type TrustCenterAccessGraphUpdateMutation$variables = {
  input: UpdateTrustCenterAccessInput;
};
export type TrustCenterAccessGraphUpdateMutation$data = {
  readonly updateTrustCenterAccess: {
    readonly trustCenterAccess: {
      readonly active: boolean;
      readonly activeCount: number;
      readonly createdAt: any;
      readonly email: string;
      readonly hasAcceptedNonDisclosureAgreement: boolean;
      readonly id: string;
      readonly lastTokenExpiresAt: any | null | undefined;
      readonly name: string;
      readonly pendingRequestCount: number;
      readonly updatedAt: any;
    };
  };
};
export type TrustCenterAccessGraphUpdateMutation = {
  response: TrustCenterAccessGraphUpdateMutation$data;
  variables: TrustCenterAccessGraphUpdateMutation$variables;
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
            "name": "active",
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
            "name": "lastTokenExpiresAt",
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
    "name": "TrustCenterAccessGraphUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "TrustCenterAccessGraphUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "8b330240bc19bbebb3244a49dc0a6db0",
    "id": null,
    "metadata": {},
    "name": "TrustCenterAccessGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterAccessGraphUpdateMutation(\n  $input: UpdateTrustCenterAccessInput!\n) {\n  updateTrustCenterAccess(input: $input) {\n    trustCenterAccess {\n      id\n      email\n      name\n      active\n      hasAcceptedNonDisclosureAgreement\n      createdAt\n      updatedAt\n      lastTokenExpiresAt\n      pendingRequestCount\n      activeCount\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "f78b4d25b0977a14269491db951a1d9d";

export default node;
