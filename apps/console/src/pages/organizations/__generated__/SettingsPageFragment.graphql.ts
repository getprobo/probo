/**
 * @generated SignedSource<<c071187d5ef9cecb128621c57f97393b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type SettingsPageFragment$data = {
  readonly connectors: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly createdAt: any;
        readonly id: string;
        readonly name: string;
        readonly type: string;
      };
    }>;
  };
  readonly description: string | null | undefined;
  readonly email: string | null | undefined;
  readonly headquarterAddress: string | null | undefined;
  readonly id: string;
  readonly logoUrl: string | null | undefined;
  readonly name: string;
  readonly users: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly createdAt: any;
        readonly email: string;
        readonly fullName: string;
        readonly id: string;
      };
    }>;
  };
  readonly websiteUrl: string | null | undefined;
  readonly " $fragmentType": "SettingsPageFragment";
};
export type SettingsPageFragment$key = {
  readonly " $data"?: SettingsPageFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"SettingsPageFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "email",
  "storageKey": null
},
v3 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 100
  }
],
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "SettingsPageFragment",
  "selections": [
    (v0/*: any*/),
    (v1/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "logoUrl",
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
      "name": "websiteUrl",
      "storageKey": null
    },
    (v2/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "headquarterAddress",
      "storageKey": null
    },
    {
      "alias": null,
      "args": (v3/*: any*/),
      "concreteType": "UserConnection",
      "kind": "LinkedField",
      "name": "users",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "UserEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "User",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v0/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "fullName",
                  "storageKey": null
                },
                (v2/*: any*/),
                (v4/*: any*/)
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "users(first:100)"
    },
    {
      "alias": null,
      "args": (v3/*: any*/),
      "concreteType": "ConnectorConnection",
      "kind": "LinkedField",
      "name": "connectors",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "ConnectorEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "Connector",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v0/*: any*/),
                (v1/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "type",
                  "storageKey": null
                },
                (v4/*: any*/)
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "connectors(first:100)"
    }
  ],
  "type": "Organization",
  "abstractKey": null
};
})();

(node as any).hash = "bdd548b7b3fa4c1cfef32569dc63d45d";

export default node;
