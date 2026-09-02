// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import type { INodeProperties, IExecuteFunctions, INodeExecutionData, IDataObject } from 'n8n-workflow';
import { NodeOperationError } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Access Review Source ID',
		name: 'accessReviewSourceId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['accessReviewSource'],
				operation: ['probe'],
			},
		},
		default: '',
		description: 'The ID of the access review source',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const accessReviewSourceId = this.getNodeParameter('accessReviewSourceId', itemIndex) as string;

	const sourceQuery = `
		query GetAccessReviewSource($id: ID!) {
			node(id: $id) {
				__typename
				... on AccessReviewSource {
					id
					connectorId
				}
			}
		}
	`;

	const sourceResponse = await proboApiRequest.call(this, sourceQuery, {
		id: accessReviewSourceId,
	});

	const node = (sourceResponse.data as IDataObject | undefined)?.node as IDataObject | undefined;
	if (!node || node.__typename !== 'AccessReviewSource') {
		throw new NodeOperationError(
			this.getNode(),
			`Access review source ${accessReviewSourceId} not found`,
		);
	}

	const connectorId = node.connectorId as string | undefined;
	if (!connectorId) {
		throw new NodeOperationError(
			this.getNode(),
			`Access review source ${accessReviewSourceId} has no connector`,
		);
	}

	const query = `
		mutation ProbeConnector($input: ProbeConnectorInput!) {
			probeConnector(input: $input) {
				ok
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, {
		input: { connectorId },
	});

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
