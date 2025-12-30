/**
 * @generated SignedSource<<879c7bec9540c7d9e46eecb2fd5243d7>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ObligationStatus = "COMPLIANT" | "NON_COMPLIANT" | "PARTIALLY_COMPLIANT";
export type ObligationType = "CONTRACTUAL" | "LEGAL";
export type UpdateObligationInput = {
  actionsToBeImplemented?: string | null | undefined;
  area?: string | null | undefined;
  dueDate?: any | null | undefined;
  id: string;
  lastReviewDate?: any | null | undefined;
  ownerId?: string | null | undefined;
  regulator?: string | null | undefined;
  requirement?: string | null | undefined;
  source?: string | null | undefined;
  status?: ObligationStatus | null | undefined;
  type?: ObligationType | null | undefined;
};
export type ObligationGraphUpdateMutation$variables = {
  input: UpdateObligationInput;
};
export type ObligationGraphUpdateMutation$data = {
  readonly updateObligation: {
    readonly obligation: {
      readonly actionsToBeImplemented: string | null | undefined;
      readonly area: string | null | undefined;
      readonly dueDate: any | null | undefined;
      readonly id: string;
      readonly lastReviewDate: any | null | undefined;
      readonly owner: {
        readonly fullName: string;
        readonly id: string;
      };
      readonly regulator: string | null | undefined;
      readonly requirement: string | null | undefined;
      readonly source: string | null | undefined;
      readonly status: ObligationStatus;
      readonly type: ObligationType;
      readonly updatedAt: any;
    };
  };
};
export type ObligationGraphUpdateMutation = {
  response: ObligationGraphUpdateMutation$data;
  variables: ObligationGraphUpdateMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v2 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "UpdateObligationPayload",
    "kind": "LinkedField",
    "name": "updateObligation",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Obligation",
        "kind": "LinkedField",
        "name": "obligation",
        "plural": false,
        "selections": [
          (v1/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "area",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "source",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "requirement",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "actionsToBeImplemented",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "regulator",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "type",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "lastReviewDate",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "dueDate",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "status",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "concreteType": "People",
            "kind": "LinkedField",
            "name": "owner",
            "plural": false,
            "selections": [
              (v1/*: any*/),
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "fullName",
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "updatedAt",
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
    "name": "ObligationGraphUpdateMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ObligationGraphUpdateMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "f9c2c73fd39c2d25eb6aa8d9e6670652",
    "id": null,
    "metadata": {},
    "name": "ObligationGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation ObligationGraphUpdateMutation(\n  $input: UpdateObligationInput!\n) {\n  updateObligation(input: $input) {\n    obligation {\n      id\n      area\n      source\n      requirement\n      actionsToBeImplemented\n      regulator\n      type\n      lastReviewDate\n      dueDate\n      status\n      owner {\n        id\n        fullName\n      }\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "07663daac0c8341a570a0bf3a41c8b77";

export default node;
