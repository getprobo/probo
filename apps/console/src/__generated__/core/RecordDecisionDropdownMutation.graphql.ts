/**
 * @generated SignedSource<<d3caf4bf2c59e4b4e520fa235e7579a6>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type AccessEntryDecision = "APPROVED" | "DEFER" | "ESCALATE" | "MODIFY" | "PENDING" | "REVOKE";
export type RecordAccessEntryDecisionInput = {
  accessEntryId: string;
  decision: AccessEntryDecision;
  decisionNote?: string | null | undefined;
};
export type RecordDecisionDropdownMutation$variables = {
  input: RecordAccessEntryDecisionInput;
};
export type RecordDecisionDropdownMutation$data = {
  readonly recordAccessEntryDecision: {
    readonly accessEntry: {
      readonly decidedAt: string | null | undefined;
      readonly decidedBy: {
        readonly fullName: string;
        readonly id: string;
      } | null | undefined;
      readonly decision: AccessEntryDecision;
      readonly decisionNote: string | null | undefined;
      readonly id: string;
    };
  };
};
export type RecordDecisionDropdownMutation = {
  response: RecordDecisionDropdownMutation$data;
  variables: RecordDecisionDropdownMutation$variables;
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
    "concreteType": "RecordAccessEntryDecisionPayload",
    "kind": "LinkedField",
    "name": "recordAccessEntryDecision",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "AccessEntry",
        "kind": "LinkedField",
        "name": "accessEntry",
        "plural": false,
        "selections": [
          (v1/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "decision",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "decidedAt",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "decisionNote",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "concreteType": "People",
            "kind": "LinkedField",
            "name": "decidedBy",
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
    "name": "RecordDecisionDropdownMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "RecordDecisionDropdownMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "fdd34e267026c1722b3204a5e4b955b9",
    "id": null,
    "metadata": {},
    "name": "RecordDecisionDropdownMutation",
    "operationKind": "mutation",
    "text": "mutation RecordDecisionDropdownMutation(\n  $input: RecordAccessEntryDecisionInput!\n) {\n  recordAccessEntryDecision(input: $input) {\n    accessEntry {\n      id\n      decision\n      decidedAt\n      decisionNote\n      decidedBy {\n        id\n        fullName\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "e755c6439e99ca32bf8827c39d04072f";

export default node;
