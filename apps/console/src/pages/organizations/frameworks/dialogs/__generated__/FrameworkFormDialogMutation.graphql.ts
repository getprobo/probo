/**
 * @generated SignedSource<<af136d6db1a014d6241ad2c9c6dd5f55>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CreateFrameworkInput = {
  description?: string | null | undefined;
  name: string;
  organizationId: string;
};
export type FrameworkFormDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateFrameworkInput;
};
export type FrameworkFormDialogMutation$data = {
  readonly createFramework: {
    readonly frameworkEdge: {
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"FrameworksPageCardFragment">;
      };
    };
  };
};
export type FrameworkFormDialogMutation = {
  response: FrameworkFormDialogMutation$data;
  variables: FrameworkFormDialogMutation$variables;
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
    "name": "FrameworkFormDialogMutation",
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
    "name": "FrameworkFormDialogMutation",
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
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "lightLogoURL",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "darkLogoURL",
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
    "cacheID": "c66177d87b59d87fdf7cbc9268af431a",
    "id": null,
    "metadata": {},
    "name": "FrameworkFormDialogMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkFormDialogMutation(\n  $input: CreateFrameworkInput!\n) {\n  createFramework(input: $input) {\n    frameworkEdge {\n      node {\n        id\n        ...FrameworksPageCardFragment\n      }\n    }\n  }\n}\n\nfragment FrameworksPageCardFragment on Framework {\n  id\n  name\n  description\n  lightLogoURL\n  darkLogoURL\n}\n"
  }
};
})();

(node as any).hash = "efed7b4ef49eea43e0bf9c0a8839c4ee";

export default node;
