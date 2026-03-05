/**
 * @generated SignedSource<<2f95db3fc1f72db6d38bed52fb1e89bd>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ProfileState = "ACTIVE" | "INACTIVE";
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
export type CompliancePageAccessListItemUpdateMutation$variables = {
  input: UpdateTrustCenterAccessInput;
};
export type CompliancePageAccessListItemUpdateMutation$data = {
  readonly updateTrustCenterAccess: {
    readonly trustCenterAccess: {
      readonly activeCount: number;
      readonly createdAt: string;
      readonly id: string;
      readonly pendingRequestCount: number;
      readonly profile: {
        readonly emailAddress: string;
        readonly fullName: string;
        readonly state: ProfileState;
      };
      readonly updatedAt: string;
    };
  };
};
export type CompliancePageAccessListItemUpdateMutation = {
  response: CompliancePageAccessListItemUpdateMutation$data;
  variables: CompliancePageAccessListItemUpdateMutation$variables;
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
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "pendingRequestCount",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "activeCount",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fullName",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "emailAddress",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "state",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "CompliancePageAccessListItemUpdateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
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
              (v2/*: any*/),
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "Profile",
                "kind": "LinkedField",
                "name": "profile",
                "plural": false,
                "selections": [
                  (v7/*: any*/),
                  (v8/*: any*/),
                  (v9/*: any*/)
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "CompliancePageAccessListItemUpdateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
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
              (v2/*: any*/),
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "Profile",
                "kind": "LinkedField",
                "name": "profile",
                "plural": false,
                "selections": [
                  (v7/*: any*/),
                  (v8/*: any*/),
                  (v9/*: any*/),
                  (v2/*: any*/)
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "f8d08c05c3a5c935c68ca7ef509cf376",
    "id": null,
    "metadata": {},
    "name": "CompliancePageAccessListItemUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageAccessListItemUpdateMutation(\n  $input: UpdateTrustCenterAccessInput!\n) {\n  updateTrustCenterAccess(input: $input) {\n    trustCenterAccess {\n      id\n      createdAt\n      updatedAt\n      pendingRequestCount\n      activeCount\n      profile {\n        fullName\n        emailAddress\n        state\n        id\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "2912ca422f9edc3806a7e1ca6cb21028";

export default node;
