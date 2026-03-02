/**
 * @generated SignedSource<<b78ad7244ad348d2622288743eff7a9b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type FrameworkDetailPageFragment$data = {
  readonly canCreateControl: boolean;
  readonly canDelete: boolean;
  readonly canExport: boolean;
  readonly canUpdate: boolean;
  readonly controls: {
    readonly __id: string;
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly bestPractice: boolean;
        readonly id: string;
        readonly name: string;
        readonly sectionTitle: string;
      };
    }>;
  };
  readonly darkLogoURL: string | null | undefined;
  readonly description: string | null | undefined;
  readonly id: string;
  readonly lightLogoURL: string | null | undefined;
  readonly name: string;
  readonly organization: {
    readonly name: string;
  };
  readonly " $fragmentType": "FrameworkDetailPageFragment";
};
export type FrameworkDetailPageFragment$key = {
  readonly " $data"?: FrameworkDetailPageFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"FrameworkDetailPageFragment">;
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
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "FrameworkDetailPageFragment",
  "selections": [
    (v0/*: any*/),
    (v1/*: any*/),
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
    },
    {
      "alias": "canExport",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:franework:export"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:franework:export\")"
    },
    {
      "alias": "canUpdate",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:framework:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:framework:update\")"
    },
    {
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:framework:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:framework:delete\")"
    },
    {
      "alias": "canCreateControl",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:control:create"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:control:create\")"
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "Organization",
      "kind": "LinkedField",
      "name": "organization",
      "plural": false,
      "selections": [
        (v1/*: any*/)
      ],
      "storageKey": null
    },
    {
      "alias": null,
      "args": [
        {
          "kind": "Literal",
          "name": "first",
          "value": 250
        },
        {
          "kind": "Literal",
          "name": "orderBy",
          "value": {
            "direction": "ASC",
            "field": "SECTION_TITLE"
          }
        }
      ],
      "concreteType": "ControlConnection",
      "kind": "LinkedField",
      "name": "controls",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "ControlEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "Control",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v0/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "sectionTitle",
                  "storageKey": null
                },
                (v1/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "bestPractice",
                  "storageKey": null
                }
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        },
        {
          "kind": "ClientExtension",
          "selections": [
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "__id",
              "storageKey": null
            }
          ]
        }
      ],
      "storageKey": "controls(first:250,orderBy:{\"direction\":\"ASC\",\"field\":\"SECTION_TITLE\"})"
    }
  ],
  "type": "Framework",
  "abstractKey": null
};
})();

(node as any).hash = "f6eba8f00d1e491f79bc201851b8c69d";

export default node;
