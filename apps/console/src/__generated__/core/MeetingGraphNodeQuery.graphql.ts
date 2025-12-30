/**
 * @generated SignedSource<<c19b786d81cd6f7adc4da4a4c3ad9346>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MeetingGraphNodeQuery$variables = {
  meetingId: string;
};
export type MeetingGraphNodeQuery$data = {
  readonly node: {
    readonly " $fragmentSpreads": FragmentRefs<"MeetingDetailPageMeetingFragment">;
  };
};
export type MeetingGraphNodeQuery = {
  response: MeetingGraphNodeQuery$data;
  variables: MeetingGraphNodeQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "meetingId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "meetingId"
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
    "name": "MeetingGraphNodeQuery",
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
            "name": "MeetingDetailPageMeetingFragment"
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
    "name": "MeetingGraphNodeQuery",
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
                  (v2/*: any*/),
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
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "a4df3c8c55dc1bf379bba19af16448d7",
    "id": null,
    "metadata": {},
    "name": "MeetingGraphNodeQuery",
    "operationKind": "query",
    "text": "query MeetingGraphNodeQuery(\n  $meetingId: ID!\n) {\n  node(id: $meetingId) {\n    __typename\n    ...MeetingDetailPageMeetingFragment\n    id\n  }\n}\n\nfragment MeetingDetailPageMeetingFragment on Meeting {\n  id\n  name\n  date\n  minutes\n  attendees {\n    id\n    fullName\n  }\n}\n"
  }
};
})();

(node as any).hash = "08844ea87fd5de5c7fe55583373e0468";

export default node;
