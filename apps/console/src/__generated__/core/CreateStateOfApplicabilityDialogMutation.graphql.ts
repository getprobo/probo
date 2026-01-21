/**
 * @generated SignedSource<<c9791128a3674bf247b23005471bdc50>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CreateStateOfApplicabilityInput = {
  name: string;
  organizationId: string;
  ownerId: string;
};
export type CreateStateOfApplicabilityDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateStateOfApplicabilityInput;
};
export type CreateStateOfApplicabilityDialogMutation$data = {
  readonly createStateOfApplicability: {
    readonly stateOfApplicabilityEdge: {
      readonly node: {
        readonly canDelete: boolean;
        readonly createdAt: string;
        readonly id: string;
        readonly name: string;
        readonly snapshotId: string | null | undefined;
        readonly sourceId: string | null | undefined;
        readonly updatedAt: string;
        readonly " $fragmentSpreads": FragmentRefs<"StateOfApplicabilityRowFragment">;
      };
    };
  };
};
export type CreateStateOfApplicabilityDialogMutation = {
  response: CreateStateOfApplicabilityDialogMutation$data;
  variables: CreateStateOfApplicabilityDialogMutation$variables;
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
  "name": "sourceId",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "snapshotId",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v9 = {
  "alias": "canDelete",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:state-of-applicability:delete"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:state-of-applicability:delete\")"
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "CreateStateOfApplicabilityDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateStateOfApplicabilityPayload",
        "kind": "LinkedField",
        "name": "createStateOfApplicability",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "StateOfApplicabilityEdge",
            "kind": "LinkedField",
            "name": "stateOfApplicabilityEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "StateOfApplicability",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/),
                  (v6/*: any*/),
                  (v7/*: any*/),
                  (v8/*: any*/),
                  (v9/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "StateOfApplicabilityRowFragment"
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
    "name": "CreateStateOfApplicabilityDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateStateOfApplicabilityPayload",
        "kind": "LinkedField",
        "name": "createStateOfApplicability",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "StateOfApplicabilityEdge",
            "kind": "LinkedField",
            "name": "stateOfApplicabilityEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "StateOfApplicability",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/),
                  (v6/*: any*/),
                  (v7/*: any*/),
                  (v8/*: any*/),
                  (v9/*: any*/),
                  {
                    "alias": "statementsInfo",
                    "args": null,
                    "concreteType": "ApplicabilityStatementConnection",
                    "kind": "LinkedField",
                    "name": "applicabilityStatements",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "totalCount",
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
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
            "name": "stateOfApplicabilityEdge",
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
    "cacheID": "cabae20b2fee98a353eaefaad0a28b4b",
    "id": null,
    "metadata": {},
    "name": "CreateStateOfApplicabilityDialogMutation",
    "operationKind": "mutation",
    "text": "mutation CreateStateOfApplicabilityDialogMutation(\n  $input: CreateStateOfApplicabilityInput!\n) {\n  createStateOfApplicability(input: $input) {\n    stateOfApplicabilityEdge {\n      node {\n        id\n        name\n        sourceId\n        snapshotId\n        createdAt\n        updatedAt\n        canDelete: permission(action: \"core:state-of-applicability:delete\")\n        ...StateOfApplicabilityRowFragment\n      }\n    }\n  }\n}\n\nfragment StateOfApplicabilityRowFragment on StateOfApplicability {\n  id\n  name\n  createdAt\n  canDelete: permission(action: \"core:state-of-applicability:delete\")\n  statementsInfo: applicabilityStatements {\n    totalCount\n  }\n}\n"
  }
};
})();

(node as any).hash = "e5a7d1dc6b0401246d646dde983bc8a0";

export default node;
