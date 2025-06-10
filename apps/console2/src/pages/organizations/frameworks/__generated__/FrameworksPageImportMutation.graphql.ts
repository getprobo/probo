/**
 * @generated SignedSource<<9436ef052ea63a9785d82ec810e810f1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ImportFrameworkInput = {
  file: any;
  organizationId: string;
};
export type FrameworksPageImportMutation$variables = {
  connections: ReadonlyArray<string>;
  input: ImportFrameworkInput;
};
export type FrameworksPageImportMutation$data = {
  readonly importFramework: {
    readonly frameworkEdge: {
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"FrameworksPageCardFragment">;
      };
    };
  };
};
export type FrameworksPageImportMutation = {
  response: FrameworksPageImportMutation$data;
  variables: FrameworksPageImportMutation$variables;
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
    "name": "FrameworksPageImportMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "ImportFrameworkPayload",
        "kind": "LinkedField",
        "name": "importFramework",
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
    "name": "FrameworksPageImportMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "ImportFrameworkPayload",
        "kind": "LinkedField",
        "name": "importFramework",
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
    "cacheID": "87a38788cb2e060657b552af69acb707",
    "id": null,
    "metadata": {},
    "name": "FrameworksPageImportMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworksPageImportMutation(\n  $input: ImportFrameworkInput!\n) {\n  importFramework(input: $input) {\n    frameworkEdge {\n      node {\n        id\n        ...FrameworksPageCardFragment\n      }\n    }\n  }\n}\n\nfragment FrameworksPageCardFragment on Framework {\n  id\n  name\n  description\n}\n"
  }
};
})();

(node as any).hash = "e2ff0f6d6780cf967a021a093542dab9";

export default node;
