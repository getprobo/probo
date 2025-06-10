/**
 * @generated SignedSource<<74725fda85d702538289d6a97ef17399>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CreateControlDocumentMappingInput = {
  controlId: string;
  documentId: string;
};
export type DocumentControlsTab_attachControlMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateControlDocumentMappingInput;
};
export type DocumentControlsTab_attachControlMutation$data = {
  readonly createControlDocumentMapping: {
    readonly controlEdge: {
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"LinkedControlsCardFragment">;
      };
    };
  };
};
export type DocumentControlsTab_attachControlMutation = {
  response: DocumentControlsTab_attachControlMutation$data;
  variables: DocumentControlsTab_attachControlMutation$variables;
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
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
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
    "name": "DocumentControlsTab_attachControlMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateControlDocumentMappingPayload",
        "kind": "LinkedField",
        "name": "createControlDocumentMapping",
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
                  (v3/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "LinkedControlsCardFragment"
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
    "name": "DocumentControlsTab_attachControlMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateControlDocumentMappingPayload",
        "kind": "LinkedField",
        "name": "createControlDocumentMapping",
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
                  (v3/*: any*/),
                  (v4/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "sectionTitle",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "Framework",
                    "kind": "LinkedField",
                    "name": "framework",
                    "plural": false,
                    "selections": [
                      (v4/*: any*/),
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
            "name": "controlEdge",
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
    "cacheID": "d283633e6a7ca3af074af6a8b716fc1c",
    "id": null,
    "metadata": {},
    "name": "DocumentControlsTab_attachControlMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentControlsTab_attachControlMutation(\n  $input: CreateControlDocumentMappingInput!\n) {\n  createControlDocumentMapping(input: $input) {\n    controlEdge {\n      node {\n        id\n        ...LinkedControlsCardFragment\n      }\n    }\n  }\n}\n\nfragment LinkedControlsCardFragment on Control {\n  id\n  name\n  sectionTitle\n  framework {\n    name\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "273f0d619a2be7df7451e7354c53315a";

export default node;
