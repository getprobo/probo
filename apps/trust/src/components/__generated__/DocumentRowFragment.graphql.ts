/**
 * @generated SignedSource<<5698d403c8907dad78a4070f710f46b4>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentAccessStatus = "GRANTED" | "REJECTED" | "REQUESTED" | "REVOKED";
import { FragmentRefs } from "relay-runtime";
export type DocumentRowFragment$data = {
  readonly access: {
    readonly id: string;
    readonly status: DocumentAccessStatus;
  } | null | undefined;
  readonly id: string;
  readonly isUserAuthorized: boolean;
  readonly title: string;
  readonly " $fragmentType": "DocumentRowFragment";
};
export type DocumentRowFragment$key = {
  readonly " $data"?: DocumentRowFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentRowFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentRowFragment",
  "selections": [
    (v0/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "title",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "isUserAuthorized",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "DocumentAccess",
      "kind": "LinkedField",
      "name": "access",
      "plural": false,
      "selections": [
        (v0/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "status",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "Document",
  "abstractKey": null
};
})();

(node as any).hash = "6839ab4ad70d571dcce4e74583723244";

export default node;
