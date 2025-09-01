/**
 * @generated SignedSource<<0f5c8a4c09725f287ec32e2366f5d829>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ContinualImprovementRegistriesPriority = "HIGH" | "LOW" | "MEDIUM";
export type ContinualImprovementRegistriesStatus = "CLOSED" | "IN_PROGRESS" | "OPEN";
export type UpdateContinualImprovementRegistryInput = {
  description?: string | null | undefined;
  id: string;
  ownerId?: string | null | undefined;
  priority?: ContinualImprovementRegistriesPriority | null | undefined;
  referenceId?: string | null | undefined;
  source?: string | null | undefined;
  status?: ContinualImprovementRegistriesStatus | null | undefined;
  targetDate?: any | null | undefined;
};
export type ContinualImprovementRegistryGraphUpdateMutation$variables = {
  input: UpdateContinualImprovementRegistryInput;
};
export type ContinualImprovementRegistryGraphUpdateMutation$data = {
  readonly updateContinualImprovementRegistry: {
    readonly continualImprovementRegistry: {
      readonly description: string | null | undefined;
      readonly id: string;
      readonly owner: {
        readonly fullName: string;
        readonly id: string;
      };
      readonly priority: ContinualImprovementRegistriesPriority;
      readonly referenceId: string;
      readonly source: string | null | undefined;
      readonly status: ContinualImprovementRegistriesStatus;
      readonly targetDate: any | null | undefined;
      readonly updatedAt: any;
    };
  };
};
export type ContinualImprovementRegistryGraphUpdateMutation = {
  response: ContinualImprovementRegistryGraphUpdateMutation$data;
  variables: ContinualImprovementRegistryGraphUpdateMutation$variables;
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
    "concreteType": "UpdateContinualImprovementRegistryPayload",
    "kind": "LinkedField",
    "name": "updateContinualImprovementRegistry",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "ContinualImprovementRegistry",
        "kind": "LinkedField",
        "name": "continualImprovementRegistry",
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
    "name": "ContinualImprovementRegistryGraphUpdateMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ContinualImprovementRegistryGraphUpdateMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "4fa7e841eea92d03f30fadf630135159",
    "id": null,
    "metadata": {},
    "name": "ContinualImprovementRegistryGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation ContinualImprovementRegistryGraphUpdateMutation(\n  $input: UpdateContinualImprovementRegistryInput!\n) {\n  updateContinualImprovementRegistry(input: $input) {\n    continualImprovementRegistry {\n      id\n      referenceId\n      description\n      source\n      targetDate\n      status\n      priority\n      owner {\n        id\n        fullName\n      }\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "82fad800ffad42813aeceda43c3304a3";

export default node;
