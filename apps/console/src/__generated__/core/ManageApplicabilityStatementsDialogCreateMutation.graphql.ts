/**
 * @generated SignedSource<<b94c40a56e3f6a02f09f329d0857f5f3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateApplicabilityStatementInput = {
  applicability: boolean;
  controlId: string;
  justification?: string | null | undefined;
  stateOfApplicabilityId: string;
};
export type ManageApplicabilityStatementsDialogCreateMutation$variables = {
  input: CreateApplicabilityStatementInput;
};
export type ManageApplicabilityStatementsDialogCreateMutation$data = {
  readonly createApplicabilityStatement: {
    readonly stateOfApplicabilityControlEdge: {
      readonly node: {
        readonly applicability: boolean;
        readonly controlId: string;
        readonly justification: string | null | undefined;
        readonly stateOfApplicabilityId: string;
      };
    };
  };
};
export type ManageApplicabilityStatementsDialogCreateMutation = {
  response: ManageApplicabilityStatementsDialogCreateMutation$data;
  variables: ManageApplicabilityStatementsDialogCreateMutation$variables;
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
  "name": "stateOfApplicabilityId",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "controlId",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "applicability",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "justification",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "ManageApplicabilityStatementsDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "CreateApplicabilityStatementPayload",
        "kind": "LinkedField",
        "name": "createApplicabilityStatement",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "StateOfApplicabilityControlEdge",
            "kind": "LinkedField",
            "name": "stateOfApplicabilityControlEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "StateOfApplicabilityControl",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v2/*: any*/),
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/)
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
    "name": "ManageApplicabilityStatementsDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "CreateApplicabilityStatementPayload",
        "kind": "LinkedField",
        "name": "createApplicabilityStatement",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "StateOfApplicabilityControlEdge",
            "kind": "LinkedField",
            "name": "stateOfApplicabilityControlEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "StateOfApplicabilityControl",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v2/*: any*/),
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "id",
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
      }
    ]
  },
  "params": {
    "cacheID": "428b4ab024585d8592798be95f6bed9e",
    "id": null,
    "metadata": {},
    "name": "ManageApplicabilityStatementsDialogCreateMutation",
    "operationKind": "mutation",
    "text": "mutation ManageApplicabilityStatementsDialogCreateMutation(\n  $input: CreateApplicabilityStatementInput!\n) {\n  createApplicabilityStatement(input: $input) {\n    stateOfApplicabilityControlEdge {\n      node {\n        stateOfApplicabilityId\n        controlId\n        applicability\n        justification\n        id\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "651f9bb0ee14cdbd644bb0be59088dce";

export default node;
