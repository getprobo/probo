/**
 * @generated SignedSource<<4744f325593bfe652352407dd9ee6e5c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CreateControlObligationMappingInput = {
  controlId: string;
  obligationId: string;
};
export type FrameworkControlPageAttachObligationMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateControlObligationMappingInput;
};
export type FrameworkControlPageAttachObligationMutation$data = {
  readonly createControlObligationMapping: {
    readonly obligationEdge: {
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"LinkedObligationsCardFragment">;
      };
    };
  };
};
export type FrameworkControlPageAttachObligationMutation = {
  response: FrameworkControlPageAttachObligationMutation$data;
  variables: FrameworkControlPageAttachObligationMutation$variables;
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
    "name": "FrameworkControlPageAttachObligationMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateControlObligationMappingPayload",
        "kind": "LinkedField",
        "name": "createControlObligationMapping",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "ObligationEdge",
            "kind": "LinkedField",
            "name": "obligationEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Obligation",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "LinkedObligationsCardFragment"
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
    "name": "FrameworkControlPageAttachObligationMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateControlObligationMappingPayload",
        "kind": "LinkedField",
        "name": "createControlObligationMapping",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "ObligationEdge",
            "kind": "LinkedField",
            "name": "obligationEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Obligation",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "requirement",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "area",
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
                    "name": "status",
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
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "fullName",
                        "storageKey": null
                      },
                      (v3/*: any*/)
                    ],
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
            "name": "obligationEdge",
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
    "cacheID": "e5dce8dc0bc03faa0e4d42696c76b59d",
    "id": null,
    "metadata": {},
    "name": "FrameworkControlPageAttachObligationMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkControlPageAttachObligationMutation(\n  $input: CreateControlObligationMappingInput!\n) {\n  createControlObligationMapping(input: $input) {\n    obligationEdge {\n      node {\n        id\n        ...LinkedObligationsCardFragment\n      }\n    }\n  }\n}\n\nfragment LinkedObligationsCardFragment on Obligation {\n  id\n  requirement\n  area\n  source\n  status\n  owner {\n    fullName\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "3c4bfe48fd9db37ce70cecd890d7248a";

export default node;
