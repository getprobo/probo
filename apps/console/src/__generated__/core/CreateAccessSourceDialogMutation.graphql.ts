/**
 * @generated SignedSource<<346f9b9e80b8897a49f3d474d015cbf2>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CreateAccessSourceInput = {
  accessReviewId: string;
  connectorId?: string | null | undefined;
  csvData?: string | null | undefined;
  name: string;
};
export type CreateAccessSourceDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateAccessSourceInput;
};
export type CreateAccessSourceDialogMutation$data = {
  readonly createAccessSource: {
    readonly accessSourceEdge: {
      readonly node: {
        readonly createdAt: string;
        readonly id: string;
        readonly name: string;
        readonly " $fragmentSpreads": FragmentRefs<"AccessSourceRowFragment">;
      };
    };
  };
};
export type CreateAccessSourceDialogMutation = {
  response: CreateAccessSourceDialogMutation$data;
  variables: CreateAccessSourceDialogMutation$variables;
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
  "name": "id",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
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
    "name": "CreateAccessSourceDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateAccessSourcePayload",
        "kind": "LinkedField",
        "name": "createAccessSource",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "AccessSourceEdge",
            "kind": "LinkedField",
            "name": "accessSourceEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "AccessSource",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "AccessSourceRowFragment"
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
    "name": "CreateAccessSourceDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateAccessSourcePayload",
        "kind": "LinkedField",
        "name": "createAccessSource",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "AccessSourceEdge",
            "kind": "LinkedField",
            "name": "accessSourceEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "AccessSource",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "connectorId",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "Connector",
                    "kind": "LinkedField",
                    "name": "connector",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "provider",
                        "storageKey": null
                      },
                      (v3/*: any*/)
                    ],
                    "storageKey": null
                  },
                  {
                    "alias": "canDelete",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:access-source:delete"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:access-source:delete\")"
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
            "name": "accessSourceEdge",
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
    "cacheID": "b05a9cbce246e1875983f7078a51c0be",
    "id": null,
    "metadata": {},
    "name": "CreateAccessSourceDialogMutation",
    "operationKind": "mutation",
    "text": "mutation CreateAccessSourceDialogMutation(\n  $input: CreateAccessSourceInput!\n) {\n  createAccessSource(input: $input) {\n    accessSourceEdge {\n      node {\n        id\n        name\n        createdAt\n        ...AccessSourceRowFragment\n      }\n    }\n  }\n}\n\nfragment AccessSourceRowFragment on AccessSource {\n  id\n  name\n  connectorId\n  connector {\n    provider\n    id\n  }\n  createdAt\n  canDelete: permission(action: \"core:access-source:delete\")\n}\n"
  }
};
})();

(node as any).hash = "4bd969a970bb943c5db18c0a11fa2756";

export default node;
