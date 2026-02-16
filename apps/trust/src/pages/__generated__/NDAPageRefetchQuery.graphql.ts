/**
 * @generated SignedSource<<47d15b007fc81de3850393c9b158726c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type NDAPageRefetchQuery$variables = {
  id: string;
};
export type NDAPageRefetchQuery$data = {
  readonly node: {
    readonly " $fragmentSpreads": FragmentRefs<"NDAPageFragment">;
  };
};
export type NDAPageRefetchQuery = {
  response: NDAPageRefetchQuery$data;
  variables: NDAPageRefetchQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "id"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "id"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "NDAPageRefetchQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "args": null,
            "kind": "FragmentSpread",
            "name": "NDAPageFragment"
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "NDAPageRefetchQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "__typename",
            "storageKey": null
          },
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "ElectronicSignature",
                "kind": "LinkedField",
                "name": "ndaSignature",
                "plural": false,
                "selections": [
                  (v2/*: any*/),
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
                    "kind": "ScalarField",
                    "name": "consentText",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "lastError",
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "type": "TrustCenter",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "1c61baa555ad2b24794ce0bf5ecc14ac",
    "id": null,
    "metadata": {},
    "name": "NDAPageRefetchQuery",
    "operationKind": "query",
    "text": "query NDAPageRefetchQuery(\n  $id: ID!\n) {\n  node(id: $id) {\n    __typename\n    ...NDAPageFragment\n    id\n  }\n}\n\nfragment NDAPageFragment on TrustCenter {\n  ndaSignature {\n    id\n    status\n    consentText\n    lastError\n  }\n  id\n}\n"
  }
};
})();

(node as any).hash = "b53a5f455508ed45124b7d09cfb573da";

export default node;
