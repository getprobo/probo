/**
 * @generated SignedSource<<43adbd5c0838293beac5a5f900fe47bd>>
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
export type AddApplicabilityStatementDialogUpdateMutation$variables = {
  input: UpdateApplicabilityStatementInput;
};
export type AddApplicabilityStatementDialogUpdateMutation$data = {
  readonly updateApplicabilityStatement: {
    readonly applicabilityStatement: {
      readonly applicability: boolean;
      readonly id: string;
      readonly justification: string;
    };
  };
};
export type AddApplicabilityStatementDialogUpdateMutation = {
  response: AddApplicabilityStatementDialogUpdateMutation$data;
  variables: AddApplicabilityStatementDialogUpdateMutation$variables;
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
        "concreteType": "ApplicabilityStatement",
        "kind": "LinkedField",
        "name": "applicabilityStatement",
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
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "AddApplicabilityStatementDialogUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "AddApplicabilityStatementDialogUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "b3f6f94a9c59775234d78ac8695211f0",
    "id": null,
    "metadata": {},
    "name": "AddApplicabilityStatementDialogUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation AddApplicabilityStatementDialogUpdateMutation(\n  $input: UpdateApplicabilityStatementInput!\n) {\n  updateApplicabilityStatement(input: $input) {\n    applicabilityStatement {\n      id\n      applicability\n      justification\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "8c801c1409e926a39c67b1c20741e501";

export default node;
