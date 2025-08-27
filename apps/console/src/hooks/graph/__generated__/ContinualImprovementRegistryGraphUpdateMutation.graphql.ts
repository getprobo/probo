/**
 * @generated SignedSource<<befa49d6faa4e4c918a2ff8a85c7fe7d>>
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
  auditId?: string | null | undefined;
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
      readonly audit: {
        readonly framework: {
          readonly id: string;
          readonly name: string;
        };
        readonly id: string;
        readonly name: string | null | undefined;
      };
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
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v3 = [
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
            "concreteType": "Audit",
            "kind": "LinkedField",
            "name": "audit",
            "plural": false,
            "selections": [
              (v1/*: any*/),
              (v2/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "Framework",
                "kind": "LinkedField",
                "name": "framework",
                "plural": false,
                "selections": [
                  (v1/*: any*/),
                  (v2/*: any*/)
                ],
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
    "selections": (v3/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ContinualImprovementRegistryGraphUpdateMutation",
    "selections": (v3/*: any*/)
  },
  "params": {
    "cacheID": "b3b3931c06838e6e60f33738be340d6c",
    "id": null,
    "metadata": {},
    "name": "ContinualImprovementRegistryGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation ContinualImprovementRegistryGraphUpdateMutation(\n  $input: UpdateContinualImprovementRegistryInput!\n) {\n  updateContinualImprovementRegistry(input: $input) {\n    continualImprovementRegistry {\n      id\n      referenceId\n      description\n      source\n      targetDate\n      status\n      priority\n      owner {\n        id\n        fullName\n      }\n      audit {\n        id\n        name\n        framework {\n          id\n          name\n        }\n      }\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "e59f20342fd5532b1815034334cd5560";

export default node;
