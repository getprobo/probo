/**
 * @generated SignedSource<<eab35c7d6cd3334dc38eae117619b5ec>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CreateTrustCenterAccessInput = {
  email: string;
  name: string;
  trustCenterId: string;
};
export type NewCompliancePageAccessDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateTrustCenterAccessInput;
};
export type NewCompliancePageAccessDialogMutation$data = {
  readonly createTrustCenterAccess: {
    readonly trustCenterAccessEdge: {
      readonly cursor: string;
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"CompliancePageAccessListItemFragment">;
      };
    };
  };
};
export type NewCompliancePageAccessDialogMutation = {
  response: NewCompliancePageAccessDialogMutation$data;
  variables: NewCompliancePageAccessDialogMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "NewCompliancePageAccessDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateTrustCenterAccessPayload",
        "kind": "LinkedField",
        "name": "createTrustCenterAccess",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "TrustCenterAccessEdge",
            "kind": "LinkedField",
            "name": "trustCenterAccessEdge",
            "plural": false,
            "selections": [
              (v3/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "TrustCenterAccess",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v4/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "CompliancePageAccessListItemFragment"
                  }
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
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "NewCompliancePageAccessDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateTrustCenterAccessPayload",
        "kind": "LinkedField",
        "name": "createTrustCenterAccess",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "TrustCenterAccessEdge",
            "kind": "LinkedField",
            "name": "trustCenterAccessEdge",
            "plural": false,
            "selections": [
              (v3/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "TrustCenterAccess",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v4/*: any*/),
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
                    "concreteType": "Profile",
                    "kind": "LinkedField",
                    "name": "profile",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "fullName",
                        "storageKey": null
                      },
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "emailAddress",
                        "storageKey": null
                      },
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "state",
                        "storageKey": null
                      },
                      (v4/*: any*/)
                    ],
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "activeCount",
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
                    "concreteType": "ElectronicSignature",
                    "kind": "LinkedField",
                    "name": "ndaSignature",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "status",
                        "storageKey": null
                      },
                      (v4/*: any*/)
                    ],
                    "storageKey": null
                  },
                  {
                    "alias": "canUpdate",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:trust-center-access:update"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:trust-center-access:update\")"
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "trustCenterAccessEdge",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "da08cb11a0a3e5668078201ca7c634ef",
    "id": null,
    "metadata": {},
    "name": "NewCompliancePageAccessDialogMutation",
    "operationKind": "mutation",
    "text": "mutation NewCompliancePageAccessDialogMutation(\n  $input: CreateTrustCenterAccessInput!\n) {\n  createTrustCenterAccess(input: $input) {\n    trustCenterAccessEdge {\n      cursor\n      node {\n        id\n        ...CompliancePageAccessListItemFragment\n      }\n    }\n  }\n}\n\nfragment CompliancePageAccessListItemFragment on TrustCenterAccess {\n  id\n  createdAt\n  profile {\n    fullName\n    emailAddress\n    state\n    id\n  }\n  activeCount\n  pendingRequestCount\n  ndaSignature {\n    status\n    id\n  }\n  canUpdate: permission(action: \"core:trust-center-access:update\")\n}\n"
  }
};
})();

(node as any).hash = "43aac1b4e3b2006dc097ef818860f541";

export default node;
