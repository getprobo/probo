/**
 * @generated SignedSource<<a91c1ee7ea4934d7d6e524f32462832a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CreateFrameworkInput = {
  description: string;
  name: string;
  organizationId: string;
};
export type CreateFrameworkDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateFrameworkInput;
};
export type CreateFrameworkDialogMutation$data = {
  readonly createFramework: {
    readonly frameworkEdge: {
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"FrameworksPageCardFragment">;
      };
    };
  };
};
export type CreateFrameworkDialogMutation = {
  response: CreateFrameworkDialogMutation$data;
  variables: CreateFrameworkDialogMutation$variables;
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
    "name": "CreateFrameworkDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateFrameworkPayload",
        "kind": "LinkedField",
        "name": "createFramework",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "FrameworkEdge",
            "kind": "LinkedField",
            "name": "frameworkEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Framework",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "FrameworksPageCardFragment"
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
    "name": "CreateFrameworkDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateFrameworkPayload",
        "kind": "LinkedField",
        "name": "createFramework",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "FrameworkEdge",
            "kind": "LinkedField",
            "name": "frameworkEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Framework",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
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
            "name": "frameworkEdge",
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
    "cacheID": "d08bc5a14bbf8d131029da890bdab850",
    "id": null,
    "metadata": {},
    "name": "CreateFrameworkDialogMutation",
    "operationKind": "mutation",
    "text": "mutation CreateFrameworkDialogMutation(\n  $input: CreateFrameworkInput!\n) {\n  createFramework(input: $input) {\n    frameworkEdge {\n      node {\n        id\n        ...FrameworksPageCardFragment\n      }\n    }\n  }\n}\n\nfragment FrameworksPageCardFragment on Framework {\n  id\n  name\n  description\n}\n"
  }
};
})();

(node as any).hash = "957d00afa55875c9a6879292e8fa7bb1";

export default node;
