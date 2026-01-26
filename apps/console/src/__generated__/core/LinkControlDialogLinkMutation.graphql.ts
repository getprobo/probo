/**
 * @generated SignedSource<<594acd2efe2323614220c1898b2e2cc9>>
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
export type LinkControlDialogLinkMutation$variables = {
  input: CreateApplicabilityStatementInput;
};
export type LinkControlDialogLinkMutation$data = {
  readonly createApplicabilityStatement: {
    readonly applicabilityStatementEdge: {
      readonly node: {
        readonly applicability: boolean;
        readonly control: {
          readonly id: string;
        };
        readonly id: string;
        readonly justification: string;
      };
    };
  };
};
export type LinkControlDialogLinkMutation = {
  response: LinkControlDialogLinkMutation$data;
  variables: LinkControlDialogLinkMutation$variables;
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
    "concreteType": "CreateApplicabilityStatementPayload",
    "kind": "LinkedField",
    "name": "createApplicabilityStatement",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "ApplicabilityStatementEdge",
        "kind": "LinkedField",
        "name": "applicabilityStatementEdge",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "ApplicabilityStatement",
            "kind": "LinkedField",
            "name": "node",
            "plural": false,
            "selections": [
              (v1/*: any*/),
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
              },
              {
                "alias": null,
                "args": null,
                "concreteType": "Control",
                "kind": "LinkedField",
                "name": "control",
                "plural": false,
                "selections": [
                  (v1/*: any*/)
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
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "LinkControlDialogLinkMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "LinkControlDialogLinkMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "bc215d6cb4d4aa2b0396f06797f6fbfe",
    "id": null,
    "metadata": {},
    "name": "LinkControlDialogLinkMutation",
    "operationKind": "mutation",
    "text": "mutation LinkControlDialogLinkMutation(\n  $input: CreateApplicabilityStatementInput!\n) {\n  createApplicabilityStatement(input: $input) {\n    applicabilityStatementEdge {\n      node {\n        id\n        applicability\n        justification\n        control {\n          id\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "0d4c01841f5e34d1c0d5517770486985";

export default node;
