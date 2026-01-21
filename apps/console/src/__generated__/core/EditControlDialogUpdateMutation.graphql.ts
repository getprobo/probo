/**
 * @generated SignedSource<<9a29546b80fe0ae4776de8cf81d600a3>>
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
export type EditControlDialogUpdateMutation$variables = {
  input: UpdateApplicabilityStatementInput;
};
export type EditControlDialogUpdateMutation$data = {
  readonly updateApplicabilityStatement: {
    readonly applicabilityStatement: {
      readonly applicability: boolean;
      readonly id: string;
      readonly justification: string;
    };
  };
};
export type EditControlDialogUpdateMutation = {
  response: EditControlDialogUpdateMutation$data;
  variables: EditControlDialogUpdateMutation$variables;
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
    "name": "EditControlDialogUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "EditControlDialogUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "eaddaf695c16db04636c4d8a79ce6346",
    "id": null,
    "metadata": {},
    "name": "EditControlDialogUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation EditControlDialogUpdateMutation(\n  $input: UpdateApplicabilityStatementInput!\n) {\n  updateApplicabilityStatement(input: $input) {\n    applicabilityStatement {\n      id\n      applicability\n      justification\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "51acae7db47e2eac1afd6a677abca126";

export default node;
