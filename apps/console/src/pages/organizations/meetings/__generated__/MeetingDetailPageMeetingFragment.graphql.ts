/**
 * @generated SignedSource<<3684f6d6c3a9746dcfcb8bb0156d2201>>
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
  readonly date: any;
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
      "alias": null,
      "args": null,
      "concreteType": "People",
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

(node as any).hash = "ae1564a2e5115bfffe20359c5288d4eb";

export default node;
