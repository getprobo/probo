/**
 * @generated SignedSource<<ab6168a4560a52cdd137c9090c37dacc>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateStateOfApplicabilityControlMappingInput = {
  applicability: boolean;
  controlId: string;
  justification?: string | null | undefined;
  stateOfApplicabilityId: string;
};
export type LinkControlDialogLinkMutation$variables = {
  input: CreateStateOfApplicabilityControlMappingInput;
};
export type LinkControlDialogLinkMutation$data = {
  readonly createStateOfApplicabilityControlMapping: {
    readonly stateOfApplicabilityControlEdge: {
      readonly node: {
        readonly applicability: boolean;
        readonly controlId: string;
        readonly justification: string | null | undefined;
        readonly stateOfApplicabilityId: string;
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
v1 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "stateOfApplicabilityId",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "controlId",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "applicability",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "justification",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "LinkControlDialogLinkMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
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
                  (v2/*: any*/),
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/)
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
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "LinkControlDialogLinkMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
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
                  (v2/*: any*/),
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "id",
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
    ]
  },
  "params": {
    "cacheID": "59e215d4f18e0429cbf813aae4db8c94",
    "id": null,
    "metadata": {},
    "name": "LinkControlDialogLinkMutation",
    "operationKind": "mutation",
    "text": "mutation LinkControlDialogLinkMutation(\n  $input: CreateStateOfApplicabilityControlMappingInput!\n) {\n  createStateOfApplicabilityControlMapping(input: $input) {\n    stateOfApplicabilityControlEdge {\n      node {\n        stateOfApplicabilityId\n        controlId\n        applicability\n        justification\n        id\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "47757e096c6beeba6be6290a2154f841";

export default node;
