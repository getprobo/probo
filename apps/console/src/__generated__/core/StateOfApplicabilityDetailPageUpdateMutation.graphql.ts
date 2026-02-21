/**
 * @generated SignedSource<<4df3ee17d608eabfe5b51673786b5d8d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateStateOfApplicabilityInput = {
  id: string;
  name?: string | null | undefined;
  ownerId?: string | null | undefined;
};
export type StateOfApplicabilityDetailPageUpdateMutation$variables = {
  input: UpdateStateOfApplicabilityInput;
};
export type StateOfApplicabilityDetailPageUpdateMutation$data = {
  readonly updateStateOfApplicability: {
    readonly stateOfApplicability: {
      readonly createdAt: string;
      readonly id: string;
      readonly name: string;
      readonly owner: {
        readonly fullName: string;
        readonly id: string;
      };
      readonly snapshotId: string | null | undefined;
      readonly sourceId: string | null | undefined;
      readonly updatedAt: string;
    };
  };
};
export type StateOfApplicabilityDetailPageUpdateMutation = {
  response: StateOfApplicabilityDetailPageUpdateMutation$data;
  variables: StateOfApplicabilityDetailPageUpdateMutation$variables;
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
    "concreteType": "UpdateStateOfApplicabilityPayload",
    "kind": "LinkedField",
    "name": "updateStateOfApplicability",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "StateOfApplicability",
        "kind": "LinkedField",
        "name": "stateOfApplicability",
        "plural": false,
        "selections": [
          (v1/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "name",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "sourceId",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "snapshotId",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "createdAt",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "updatedAt",
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
    "name": "StateOfApplicabilityDetailPageUpdateMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "StateOfApplicabilityDetailPageUpdateMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "0236ac567b08f60ca0773d91fb1b2155",
    "id": null,
    "metadata": {},
    "name": "StateOfApplicabilityDetailPageUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation StateOfApplicabilityDetailPageUpdateMutation(\n  $input: UpdateStateOfApplicabilityInput!\n) {\n  updateStateOfApplicability(input: $input) {\n    stateOfApplicability {\n      id\n      name\n      sourceId\n      snapshotId\n      createdAt\n      updatedAt\n      owner {\n        id\n        fullName\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "0213db8aa925819738cea45aa8603d6d";

export default node;
