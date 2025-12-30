/**
 * @generated SignedSource<<6798ad28a0faeedef40b1b0a89f7563e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CreateStateOfApplicabilityControlMappingInput = {
  applicability: boolean;
  controlId: string;
  justification?: string | null | undefined;
  stateOfApplicabilityId: string;
};
export type FrameworkControlPageAttachStateOfApplicabilityMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateStateOfApplicabilityControlMappingInput;
};
export type FrameworkControlPageAttachStateOfApplicabilityMutation$data = {
  readonly createStateOfApplicabilityControlMapping: {
    readonly stateOfApplicabilityControlEdge: {
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"LinkedStatesOfApplicabilityCardFragment">;
      };
    };
  };
};
export type FrameworkControlPageAttachStateOfApplicabilityMutation = {
  response: FrameworkControlPageAttachStateOfApplicabilityMutation$data;
  variables: FrameworkControlPageAttachStateOfApplicabilityMutation$variables;
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
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "FrameworkControlPageAttachStateOfApplicabilityMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateStateOfApplicabilityControlMappingPayload",
        "kind": "LinkedField",
        "name": "createStateOfApplicabilityControlMapping",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "StateOfApplicabilityControlEdge",
            "kind": "LinkedField",
            "name": "stateOfApplicabilityControlEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "StateOfApplicabilityControl",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "LinkedStatesOfApplicabilityCardFragment"
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
    "name": "FrameworkControlPageAttachStateOfApplicabilityMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateStateOfApplicabilityControlMappingPayload",
        "kind": "LinkedField",
        "name": "createStateOfApplicabilityControlMapping",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "StateOfApplicabilityControlEdge",
            "kind": "LinkedField",
            "name": "stateOfApplicabilityControlEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "StateOfApplicabilityControl",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "stateOfApplicabilityId",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "controlId",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "StateOfApplicability",
                    "kind": "LinkedField",
                    "name": "stateOfApplicability",
                    "plural": false,
                    "selections": [
                      (v3/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "name",
                        "storageKey": null
                      }
                    ],
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
          },
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "stateOfApplicabilityControlEdge",
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
    "cacheID": "205bc4f133e268c943ef9df9881350f0",
    "id": null,
    "metadata": {},
    "name": "FrameworkControlPageAttachStateOfApplicabilityMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkControlPageAttachStateOfApplicabilityMutation(\n  $input: CreateStateOfApplicabilityControlMappingInput!\n) {\n  createStateOfApplicabilityControlMapping(input: $input) {\n    stateOfApplicabilityControlEdge {\n      node {\n        id\n        ...LinkedStatesOfApplicabilityCardFragment\n      }\n    }\n  }\n}\n\nfragment LinkedStatesOfApplicabilityCardFragment on StateOfApplicabilityControl {\n  id\n  stateOfApplicabilityId\n  controlId\n  stateOfApplicability {\n    id\n    name\n  }\n  applicability\n  justification\n}\n"
  }
};
})();

(node as any).hash = "8da7bd1e22eb2eaf857a4e97cea47e90";

export default node;
