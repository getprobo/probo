/**
 * @generated SignedSource<<86c2716d4118ff957765ad2e0577fb05>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type NonconformityStatus = "CLOSED" | "IN_PROGRESS" | "OPEN";
export type UpdateNonconformityInput = {
  auditId?: string | null | undefined;
  correctiveAction?: string | null | undefined;
  dateIdentified?: string | null | undefined;
  description?: string | null | undefined;
  dueDate?: string | null | undefined;
  effectivenessCheck?: string | null | undefined;
  id: string;
  ownerId?: string | null | undefined;
  referenceId?: string | null | undefined;
  rootCause?: string | null | undefined;
  status?: NonconformityStatus | null | undefined;
};
export type NonconformityGraphUpdateMutation$variables = {
  input: UpdateNonconformityInput;
};
export type NonconformityGraphUpdateMutation$data = {
  readonly updateNonconformity: {
    readonly nonconformity: {
      readonly audit: {
        readonly framework: {
          readonly id: string;
          readonly name: string;
        };
        readonly id: string;
      } | null | undefined;
      readonly correctiveAction: string | null | undefined;
      readonly dateIdentified: string | null | undefined;
      readonly description: string | null | undefined;
      readonly dueDate: string | null | undefined;
      readonly effectivenessCheck: string | null | undefined;
      readonly id: string;
      readonly owner: {
        readonly fullName: string;
        readonly id: string;
      };
      readonly referenceId: string;
      readonly rootCause: string;
      readonly status: NonconformityStatus;
      readonly updatedAt: string;
    };
  };
};
export type NonconformityGraphUpdateMutation = {
  response: NonconformityGraphUpdateMutation$data;
  variables: NonconformityGraphUpdateMutation$variables;
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
    "concreteType": "UpdateNonconformityPayload",
    "kind": "LinkedField",
    "name": "updateNonconformity",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Nonconformity",
        "kind": "LinkedField",
        "name": "nonconformity",
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
            "name": "dateIdentified",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "rootCause",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "correctiveAction",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "dueDate",
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
            "name": "effectivenessCheck",
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
            "concreteType": "Audit",
            "kind": "LinkedField",
            "name": "audit",
            "plural": false,
            "selections": [
              (v1/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "Framework",
                "kind": "LinkedField",
                "name": "framework",
                "plural": false,
                "selections": [
                  (v1/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "name",
                    "storageKey": null
                  }
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
    "name": "NonconformityGraphUpdateMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "NonconformityGraphUpdateMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "db365ab97fb067ad2e9b1daa68f64ed9",
    "id": null,
    "metadata": {},
    "name": "NonconformityGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation NonconformityGraphUpdateMutation(\n  $input: UpdateNonconformityInput!\n) {\n  updateNonconformity(input: $input) {\n    nonconformity {\n      id\n      referenceId\n      description\n      dateIdentified\n      rootCause\n      correctiveAction\n      dueDate\n      status\n      effectivenessCheck\n      owner {\n        id\n        fullName\n      }\n      audit {\n        id\n        framework {\n          id\n          name\n        }\n      }\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "5803de6f6f017148cca1e3066a6b3967";

export default node;
