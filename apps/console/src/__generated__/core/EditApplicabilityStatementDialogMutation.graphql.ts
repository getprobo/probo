/**
 * @generated SignedSource<<9846bbc2aaec7a7de2f57298c56b6849>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateApplicabilityStatementInput = {
  applicability: boolean;
  applicabilityStatementId: string;
  justification?: string | null | undefined;
};
export type EditApplicabilityStatementDialogMutation$variables = {
  input: UpdateApplicabilityStatementInput;
};
export type EditApplicabilityStatementDialogMutation$data = {
  readonly updateApplicabilityStatement: {
    readonly stateOfApplicabilityControlEdge: {
      readonly node: {
        readonly applicability: boolean;
        readonly controlId: string;
        readonly id: string;
        readonly justification: string | null | undefined;
        readonly stateOfApplicabilityId: string;
      };
    };
  };
};
export type EditApplicabilityStatementDialogMutation = {
  response: EditApplicabilityStatementDialogMutation$data;
  variables: EditApplicabilityStatementDialogMutation$variables;
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
    "concreteType": "UpdateApplicabilityStatementPayload",
    "kind": "LinkedField",
    "name": "updateApplicabilityStatement",
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
                "name": "stateOfApplicabilityId",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "controlId",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "applicability",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "justification",
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
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "EditApplicabilityStatementDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "EditApplicabilityStatementDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "9584d97e072ebd014225bb6756ad6b84",
    "id": null,
    "metadata": {},
    "name": "EditApplicabilityStatementDialogMutation",
    "operationKind": "mutation",
    "text": "mutation EditApplicabilityStatementDialogMutation(\n  $input: UpdateApplicabilityStatementInput!\n) {\n  updateApplicabilityStatement(input: $input) {\n    stateOfApplicabilityControlEdge {\n      node {\n        id\n        stateOfApplicabilityId\n        controlId\n        applicability\n        justification\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "def817ce2c894a1f5318c3cb14080922";

export default node;
