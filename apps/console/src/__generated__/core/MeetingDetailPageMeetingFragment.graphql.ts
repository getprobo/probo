/**
 * @generated SignedSource<<6804811557c4369cfc98c1d18ce6a668>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MeetingDetailPageMeetingFragment$data = {
  readonly attendees: ReadonlyArray<{
    readonly fullName: string;
    readonly id: string;
  }>;
  readonly canDelete: boolean;
  readonly canUpdate: boolean;
  readonly date: string;
  readonly id: string;
  readonly minutes: string | null | undefined;
  readonly name: string;
  readonly " $fragmentType": "MeetingDetailPageMeetingFragment";
};
export type MeetingDetailPageMeetingFragment$key = {
  readonly " $data"?: MeetingDetailPageMeetingFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"MeetingDetailPageMeetingFragment">;
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
  "name": "MeetingDetailPageMeetingFragment",
  "selections": [
    (v0/*: any*/),
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
      "name": "date",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "minutes",
      "storageKey": null
    },
    {
      "alias": "canUpdate",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:meeting:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:meeting:update\")"
    },
    {
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:meeting:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:meeting:delete\")"
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "Profile",
      "kind": "LinkedField",
      "name": "attendees",
      "plural": true,
      "selections": [
        (v0/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "fullName",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "Meeting",
  "abstractKey": null
};
})();

(node as any).hash = "16fcd4d8cc7dbe80bc7fd37ac5e824e8";

export default node;
