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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { organizationLocator, personLocator } from '../../locators';
import { proboApiRequest, readResourceId } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	organizationLocator({
		resource: ['device'],
		operation: ['create'],
	}),
	personLocator({
		displayName: 'Owner',
		resource: ['device'],
		operation: ['create'],
		required: false,
	}),
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = readResourceId(this.getNodeParameter('organizationId', itemIndex));
	const ownerId = readResourceId(this.getNodeParameter('ownerId', itemIndex));

	const input: { organizationId: string; ownerId?: string } = {
		organizationId,
	};

	if (ownerId) {
		input.ownerId = ownerId;
	}

	const query = `
		mutation CreateDevice($input: CreateDeviceInput!) {
			createDevice(input: $input) {
				device {
					id
					state
				}
				enrollmentToken
				serverUrl
				enrollmentUrl
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
