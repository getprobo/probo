/**
 * @generated SignedSource<<4f5d3e4c56b185607505d82119176eb5>>
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
export type AddApplicabilityStatementDialogCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateApplicabilityStatementInput;
};
export type AddApplicabilityStatementDialogCreateMutation$data = {
  readonly createApplicabilityStatement: {
    readonly applicabilityStatementEdge: {
      readonly node: {
        readonly applicability: boolean;
        readonly control: {
          readonly bestPractice: boolean;
          readonly contractual: boolean;
          readonly framework: {
            readonly id: string;
            readonly name: string;
          };
          readonly id: string;
          readonly name: string;
          readonly organization: {
            readonly id: string;
          } | null | undefined;
          readonly regulatory: boolean;
          readonly riskAssessment: boolean;
          readonly sectionTitle: string;
        };
        readonly id: string;
        readonly justification: string;
      };
    };
  };
};
export type AddApplicabilityStatementDialogCreateMutation = {
  response: AddApplicabilityStatementDialogCreateMutation$data;
  variables: AddApplicabilityStatementDialogCreateMutation$variables;
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
        (v3/*: any*/),
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
            (v3/*: any*/),
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "sectionTitle",
              "storageKey": null
            },
            (v4/*: any*/),
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "bestPractice",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "regulatory",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "contractual",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "riskAssessment",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "concreteType": "Framework",
              "kind": "LinkedField",
              "name": "framework",
              "plural": false,
              "selections": [
                (v3/*: any*/),
                (v4/*: any*/)
              ],
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "concreteType": "Organization",
              "kind": "LinkedField",
              "name": "organization",
              "plural": false,
              "selections": [
                (v3/*: any*/)
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
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "AddApplicabilityStatementDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateApplicabilityStatementPayload",
        "kind": "LinkedField",
        "name": "createApplicabilityStatement",
        "plural": false,
        "selections": [
          (v5/*: any*/)
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
    "name": "AddApplicabilityStatementDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateApplicabilityStatementPayload",
        "kind": "LinkedField",
        "name": "createApplicabilityStatement",
        "plural": false,
        "selections": [
          (v5/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "appendEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "applicabilityStatementEdge",
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
    "cacheID": "df832d2f8d480328039fdcf85bf9187c",
    "id": null,
    "metadata": {},
    "name": "AddApplicabilityStatementDialogCreateMutation",
    "operationKind": "mutation",
    "text": "mutation AddApplicabilityStatementDialogCreateMutation(\n  $input: CreateApplicabilityStatementInput!\n) {\n  createApplicabilityStatement(input: $input) {\n    applicabilityStatementEdge {\n      node {\n        id\n        applicability\n        justification\n        control {\n          id\n          sectionTitle\n          name\n          bestPractice\n          regulatory\n          contractual\n          riskAssessment\n          framework {\n            id\n            name\n          }\n          organization {\n            id\n          }\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "ddf10370932006893422c8cfb47e5587";

export default node;
