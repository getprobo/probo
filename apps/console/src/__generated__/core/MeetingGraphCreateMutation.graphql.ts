/**
 * @generated SignedSource<<cd06e77586ebdd55649cf07671903b0a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateMeetingInput = {
  attendeeIds?: ReadonlyArray<string> | null | undefined;
  date: string;
  minutes?: string | null | undefined;
  name: string;
  organizationId: string;
};
export type MeetingGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateMeetingInput;
};
export type MeetingGraphCreateMutation$data = {
  readonly createMeeting: {
    readonly meetingEdge: {
      readonly node: {
        readonly attendees: ReadonlyArray<{
          readonly fullName: string;
          readonly id: string;
        }>;
        readonly date: string;
        readonly id: string;
        readonly minutes: string | null | undefined;
        readonly name: string;
      };
    };
  };
};
export type MeetingGraphCreateMutation = {
  response: MeetingGraphCreateMutation$data;
  variables: MeetingGraphCreateMutation$variables;
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
},
v4 = {
  "alias": null,
  "args": null,
  "concreteType": "MeetingEdge",
  "kind": "LinkedField",
  "name": "meetingEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "Meeting",
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
          "concreteType": "Profile",
          "kind": "LinkedField",
          "name": "attendees",
          "plural": true,
          "selections": [
            (v3/*: any*/),
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
      "storageKey": null
    }
  ],
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
    "name": "MeetingGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateMeetingPayload",
        "kind": "LinkedField",
        "name": "createMeeting",
        "plural": false,
        "selections": [
          (v4/*: any*/)
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
    "name": "MeetingGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateMeetingPayload",
        "kind": "LinkedField",
        "name": "createMeeting",
        "plural": false,
        "selections": [
          (v4/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "meetingEdge",
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
    "cacheID": "1d0c66e942482f999e3b986af16df678",
    "id": null,
    "metadata": {},
    "name": "MeetingGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation MeetingGraphCreateMutation(\n  $input: CreateMeetingInput!\n) {\n  createMeeting(input: $input) {\n    meetingEdge {\n      node {\n        id\n        name\n        date\n        minutes\n        attendees {\n          id\n          fullName\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "27fd6781022144fe6fd247b7e2ce1aaf";

export default node;
