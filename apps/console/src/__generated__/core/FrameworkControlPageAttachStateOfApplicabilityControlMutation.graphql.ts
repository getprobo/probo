/**
 * @generated SignedSource<<9426f39407442c22140c6cfdf6798b60>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CreateApplicabilityStatementInput = {
  applicability: boolean;
  controlId: string;
  justification?: string | null | undefined;
  stateOfApplicabilityId: string;
};
export type FrameworkControlPageAttachStateOfApplicabilityControlMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateApplicabilityStatementInput;
};
export type FrameworkControlPageAttachStateOfApplicabilityControlMutation$data = {
  readonly createApplicabilityStatement: {
    readonly stateOfApplicabilityControlEdge: {
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"ControlApplicabilityStatementsCardFragment">;
      };
    };
  };
};
export type FrameworkControlPageAttachStateOfApplicabilityControlMutation = {
  response: FrameworkControlPageAttachStateOfApplicabilityControlMutation$data;
  variables: FrameworkControlPageAttachStateOfApplicabilityControlMutation$variables;
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
    "name": "FrameworkControlPageAttachStateOfApplicabilityControlMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateApplicabilityStatementPayload",
        "kind": "LinkedField",
        "name": "createApplicabilityStatement",
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
                    "name": "ControlApplicabilityStatementsCardFragment"
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
    "name": "FrameworkControlPageAttachStateOfApplicabilityControlMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateApplicabilityStatementPayload",
        "kind": "LinkedField",
        "name": "createApplicabilityStatement",
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
    "cacheID": "d5c7e62957daff31a06479103915f0ae",
    "id": null,
    "metadata": {},
    "name": "FrameworkControlPageAttachStateOfApplicabilityControlMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkControlPageAttachStateOfApplicabilityControlMutation(\n  $input: CreateApplicabilityStatementInput!\n) {\n  createApplicabilityStatement(input: $input) {\n    stateOfApplicabilityControlEdge {\n      node {\n        id\n        ...ControlApplicabilityStatementsCardFragment\n      }\n    }\n  }\n}\n\nfragment ControlApplicabilityStatementsCardFragment on StateOfApplicabilityControl {\n  id\n  stateOfApplicabilityId\n  controlId\n  stateOfApplicability {\n    id\n    name\n  }\n  applicability\n  justification\n}\n"
  }
};
})();

(node as any).hash = "ef32d913dccd88dfc539367de5a80e27";

export default node;
