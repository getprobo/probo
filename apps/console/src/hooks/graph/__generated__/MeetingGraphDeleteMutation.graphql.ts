/**
 * @generated SignedSource<<f903f8daaf7c28c3d165a215c98e05b6>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteMeetingInput = {
  meetingId: string;
};
export type MeetingGraphDeleteMutation$variables = {
  input: DeleteMeetingInput;
};
export type MeetingGraphDeleteMutation$data = {
  readonly deleteMeeting: {
    readonly deletedMeetingId: string;
  };
};
export type MeetingGraphDeleteMutation = {
  response: MeetingGraphDeleteMutation$data;
  variables: MeetingGraphDeleteMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "deletedMeetingId",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "MeetingGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "DeleteMeetingPayload",
        "kind": "LinkedField",
        "name": "deleteMeeting",
        "plural": false,
        "selections": [
          (v2/*: any*/)
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "MeetingGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "DeleteMeetingPayload",
        "kind": "LinkedField",
        "name": "deleteMeeting",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteRecord",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedMeetingId"
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "f88bdc6e79cef363e7c2aaf91dce1d16",
    "id": null,
    "metadata": {},
    "name": "MeetingGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation MeetingGraphDeleteMutation(\n  $input: DeleteMeetingInput!\n) {\n  deleteMeeting(input: $input) {\n    deletedMeetingId\n  }\n}\n"
  }
};
})();

(node as any).hash = "4d70861c6a4d4b06c753863cba054fc8";

export default node;
