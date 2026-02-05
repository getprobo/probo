/**
 * @generated SignedSource<<3fe464d047c80b94afdf64dc9cf2e23c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ContinualImprovementPriority = "HIGH" | "LOW" | "MEDIUM";
export type ContinualImprovementStatus = "CLOSED" | "IN_PROGRESS" | "OPEN";
export type UpdateContinualImprovementInput = {
  description?: string | null | undefined;
  id: string;
  ownerId?: string | null | undefined;
  priority?: ContinualImprovementPriority | null | undefined;
  referenceId?: string | null | undefined;
  source?: string | null | undefined;
  status?: ContinualImprovementStatus | null | undefined;
  targetDate?: string | null | undefined;
};
export type ContinualImprovementGraphUpdateMutation$variables = {
  input: UpdateContinualImprovementInput;
};
export type ContinualImprovementGraphUpdateMutation$data = {
  readonly updateContinualImprovement: {
    readonly continualImprovement: {
      readonly description: string | null | undefined;
      readonly id: string;
      readonly owner: {
        readonly fullName: string;
        readonly id: string;
      };
      readonly priority: ContinualImprovementPriority;
      readonly referenceId: string;
      readonly source: string | null | undefined;
      readonly status: ContinualImprovementStatus;
      readonly targetDate: string | null | undefined;
      readonly updatedAt: string;
    };
  };
};
export type ContinualImprovementGraphUpdateMutation = {
  response: ContinualImprovementGraphUpdateMutation$data;
  variables: ContinualImprovementGraphUpdateMutation$variables;
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
    "concreteType": "UpdateContinualImprovementPayload",
    "kind": "LinkedField",
    "name": "updateContinualImprovement",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "ContinualImprovement",
        "kind": "LinkedField",
        "name": "continualImprovement",
        "plural": false,
        "selections": [
          (v1/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "referenceId",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "description",
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
            "name": "targetDate",
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
            "kind": "ScalarField",
            "name": "priority",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "concreteType": "Profile",
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
    "name": "ContinualImprovementGraphUpdateMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ContinualImprovementGraphUpdateMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "437d2ec2228356103e851f53d9edb577",
    "id": null,
    "metadata": {},
    "name": "ContinualImprovementGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation ContinualImprovementGraphUpdateMutation(\n  $input: UpdateContinualImprovementInput!\n) {\n  updateContinualImprovement(input: $input) {\n    continualImprovement {\n      id\n      referenceId\n      description\n      source\n      targetDate\n      status\n      priority\n      owner {\n        id\n        fullName\n      }\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "75ef82d455a038a7c60852b7e849284d";

export default node;
