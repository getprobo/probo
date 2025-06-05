/**
 * @generated SignedSource<<1c5fe5964119705dcc3c94624cb78c8b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateControlInput = {
  description: string;
  frameworkId: string;
  name: string;
  organizationId: string;
  sectionTitle: string;
};
export type NewControlViewCreateControlMutation$variables = {
  input: CreateControlInput;
};
export type NewControlViewCreateControlMutation$data = {
  readonly createControl: {
    readonly controlEdge: {
      readonly node: {
        readonly description: string;
        readonly id: string;
        readonly name: string;
        readonly sectionTitle: string;
      };
    };
  };
};
export type NewControlViewCreateControlMutation = {
  response: NewControlViewCreateControlMutation$data;
  variables: NewControlViewCreateControlMutation$variables;
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
    "concreteType": "CreateControlPayload",
    "kind": "LinkedField",
    "name": "createControl",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "ControlEdge",
        "kind": "LinkedField",
        "name": "controlEdge",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "Control",
            "kind": "LinkedField",
            "name": "node",
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
                "name": "name",
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
                "name": "sectionTitle",
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
    "name": "NewControlViewCreateControlMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "NewControlViewCreateControlMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "2005ccdbeac4998dc5bea71013f4cb5d",
    "id": null,
    "metadata": {},
    "name": "NewControlViewCreateControlMutation",
    "operationKind": "mutation",
    "text": "mutation NewControlViewCreateControlMutation(\n  $input: CreateControlInput!\n) {\n  createControl(input: $input) {\n    controlEdge {\n      node {\n        id\n        name\n        description\n        sectionTitle\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "354415d46fd5d1b15c029b96452054d3";

export default node;
