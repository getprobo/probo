/**
 * @generated SignedSource<<6b38711542a158ba9424c744d217bed5>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type FrameworkDetailPageFragment$data = {
  readonly controls: {
    readonly __id: string;
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly id: string;
        readonly name: string;
        readonly sectionTitle: string;
      };
    }>;
  };
  readonly description: string;
  readonly id: string;
  readonly name: string;
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
      "args": [
        {
          "kind": "Literal",
          "name": "first",
          "value": 100
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
                (v1/*: any*/)
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
      "storageKey": "controls(first:100)"
    }
  ],
  "type": "Framework",
  "abstractKey": null
};
})();

(node as any).hash = "e95f82575c72a8b95c87ecd3ffc676ed";

export default node;
