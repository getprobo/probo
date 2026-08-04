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

import type {
	IDataObject,
	IExecuteFunctions,
	INodeExecutionData,
	INodeProperties,
} from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';
import { incidentFields, incidentIdField } from './fields';

export const description: INodeProperties[] = [
	incidentIdField('transition'),
	{
		displayName: 'Target Status',
		name: 'toStatus',
		type: 'options',
		displayOptions: {
			show: { resource: ['malaysiaPDPABreach'], operation: ['transition'] },
		},
		options: [
			{ name: 'Open', value: 'OPEN' },
			{ name: 'Assessing', value: 'ASSESSING' },
			{ name: 'Contained', value: 'CONTAINED' },
			{ name: 'Closed', value: 'CLOSED' },
		],
		default: 'ASSESSING',
		description: 'Target status; invalid workflow transitions are rejected',
		required: true,
	},
	{
		displayName: 'Reason',
		name: 'reason',
		type: 'string',
		displayOptions: {
			show: { resource: ['malaysiaPDPABreach'], operation: ['transition'] },
		},
		default: '',
		description: 'Reason recorded in the immutable status history',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const incidentId = this.getNodeParameter('incidentId', itemIndex) as string;
	const toStatus = this.getNodeParameter('toStatus', itemIndex) as string;
	const reason = this.getNodeParameter('reason', itemIndex, '') as string;
	const query = `
		mutation TransitionMalaysiaPDPABreachStatus($input: TransitionMalaysiaPDPABreachStatusInput!) {
			transitionMalaysiaPDPABreachStatus(input: $input) {
				incident { ${incidentFields} }
				historyEdge { node { id incidentId fromStatus toStatus changedByProfileId reason createdAt } }
			}
		}
	`;
	const input: IDataObject = { id: incidentId, toStatus };
	if (reason) input.reason = reason;
	const responseData = await proboApiRequest.call(this, query, { input });
	return { json: responseData, pairedItem: { item: itemIndex } };
}
