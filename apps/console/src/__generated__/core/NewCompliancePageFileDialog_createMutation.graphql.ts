/**
 * @generated SignedSource<<50c909a6a1d81679625acf045c6ca1c2>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type TrustCenterVisibility = "NONE" | "PRIVATE" | "PUBLIC";
export type CreateTrustCenterFileInput = {
  category: string;
  file: any;
  name: string;
  organizationId: string;
  trustCenterVisibility: TrustCenterVisibility;
};
export type NewCompliancePageFileDialog_createMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateTrustCenterFileInput;
};
export type NewCompliancePageFileDialog_createMutation$data = {
  readonly createTrustCenterFile: {
    readonly trustCenterFileEdge: {
      readonly node: {
        readonly " $fragmentSpreads": FragmentRefs<"CompliancePageFileListItem_fileFragment">;
      };
    };
  };
};
export type NewCompliancePageFileDialog_createMutation = {
  response: NewCompliancePageFileDialog_createMutation$data;
  variables: NewCompliancePageFileDialog_createMutation$variables;
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
];
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "NewCompliancePageFileDialog_createMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateTrustCenterFilePayload",
        "kind": "LinkedField",
        "name": "createTrustCenterFile",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "TrustCenterFileEdge",
            "kind": "LinkedField",
            "name": "trustCenterFileEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "TrustCenterFile",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "CompliancePageFileListItem_fileFragment"
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
    "name": "NewCompliancePageFileDialog_createMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateTrustCenterFilePayload",
        "kind": "LinkedField",
        "name": "createTrustCenterFile",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "TrustCenterFileEdge",
            "kind": "LinkedField",
            "name": "trustCenterFileEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "TrustCenterFile",
                "kind": "LinkedField",
                "name": "node",
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
                    "name": "name",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "category",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "fileUrl",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "trustCenterVisibility",
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
                    "alias": "canUpdate",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:trust-center-file:update"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:trust-center-file:update\")"
                  },
                  {
                    "alias": "canDelete",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:trust-center-file:delete"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:trust-center-file:delete\")"
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
            "name": "trustCenterFileEdge",
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
    "cacheID": "337ae413f581adb0256b0b47f0514f43",
    "id": null,
    "metadata": {},
    "name": "NewCompliancePageFileDialog_createMutation",
    "operationKind": "mutation",
    "text": "mutation NewCompliancePageFileDialog_createMutation(\n  $input: CreateTrustCenterFileInput!\n) {\n  createTrustCenterFile(input: $input) {\n    trustCenterFileEdge {\n      node {\n        ...CompliancePageFileListItem_fileFragment\n        id\n      }\n    }\n  }\n}\n\nfragment CompliancePageFileListItem_fileFragment on TrustCenterFile {\n  id\n  name\n  category\n  fileUrl\n  trustCenterVisibility\n  createdAt\n  canUpdate: permission(action: \"core:trust-center-file:update\")\n  canDelete: permission(action: \"core:trust-center-file:delete\")\n}\n"
  }
};
})();

(node as any).hash = "826c457285dd9890c39b9c11e600316f";

export default node;
